package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"
	"peer/helpers"
	"slices"
	"strconv"
	"sync"
)

type Peer struct {
	listenport int
	id         string
	conns      map[string]net.Conn
	// output serializes logging to stdout.
	output chan string
	// received forwards messages to the demo WaitGroup.
	received chan Message
	lock     sync.Mutex
	// sendLock prevents interleaved writes on a connection.
	sendLock sync.Mutex
	// Map of messages this has sent on (key is message ID).
	msgHistory   map[string]MessageHistory
	floodingLock sync.Mutex
}

type MessageHistory struct {
	content      *Message // The actual message.
	receivedFrom []string // List of IDs from all peers this peer has received this message from.
	isSent       bool     // Has this peer already sent this message.
}

func NewMessageHistory(msg *Message) *MessageHistory {
	hist := new(MessageHistory)
	hist.content = msg
	hist.receivedFrom = make([]string, 0)
	hist.isSent = false
	return hist
}

// Create a new Peer object.
func NewPeer(listenport int) *Peer {
	peer := new(Peer)
	peer.id = strconv.Itoa(listenport)
	peer.listenport = listenport
	peer.conns = make(map[string]net.Conn)
	peer.output = make(chan string, 32)
	peer.received = make(chan Message, 32)
	peer.msgHistory = make(map[string]MessageHistory)
	return peer
}

func (p *Peer) StartWithConnection(addr string, port int) {
	p.Start()
	peers, err := p.Connect(addr, port)
	if err != nil { // If connection fails, then there is no network.
		return // Do not propagate error. This peer starts its own network, so OK.
	}
	// Otherwise, a connection message was sent, and remaining connections can be established.
	p.joinNetwork(addr, peers)
}

// Start listening on the port.
func (p *Peer) Start() error {
	// Listen for connection.
	listener, err := net.Listen(helpers.PROTOCOL, ":"+strconv.Itoa(p.listenport))
	if err != nil {
		return err
	}
	// Add self to map of connections.
	p.conns[p.id] = nil
	// Goroutine to listen to any number of peers.
	go p.listenForPeers(listener)
	// Print logs from a single goroutine to avoid interleaving.
	go p.printOutput()
	return nil
}

// Listen for peers. When connection is established, prepare communication with them.
func (p *Peer) listenForPeers(listener net.Listener) {
	// Defer ensures listener is closed when this goroutine returns.
	defer listener.Close()
	for {
		// Wait to establish connection.
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		// Prepare network communication with the new connection.
		p.prepareConnection(conn)
	}
}

// Prepare communication with a peer.
func (p *Peer) prepareConnection(conn net.Conn) ([]string, error) {
	reader := bufio.NewReader(conn)
	// Handshake: announce our id, then learn the peer id.
	p.lock.Lock() // Lock to ensure no messages get sent to this before it has had a chance to update decoders map.
	defer p.lock.Unlock()
	// Send a connect message. The other peer will receivee it in their prepareConnection at the readMessage line.
	// Message payload is the list of known peers, this knows about (self, if this is a new peer, otherwise the entire network).
	payload, _ := json.Marshal(slices.Collect(maps.Keys(p.conns))) // Convert connection keys to []string and marshal it.
	_ = p.writeMessage(conn, &Message{Type: helpers.CONNECT_MESSAGE_TYPE, From: p.id, Payload: payload})
	// Wait for reply to establish their name.
	msg, err := p.readMessage(reader)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	p.conns[msg.From] = conn
	go p.handleDecode(msg.From, conn, reader)
	// Unmarshal the received list of peers the connection knew about.
	var peers []string
	json.Unmarshal(msg.Payload, &peers)
	return peers, nil
}

// Wait for messages from connection.
func (p *Peer) handleDecode(peerID string, conn net.Conn, reader *bufio.Reader) {
	// Defer ensures cleanup when the reader loop ends.
	defer func() {
		_ = conn.Close()
		p.lock.Lock()
		delete(p.conns, peerID)
		p.lock.Unlock()
	}()
	for {
		// Wait until receving a message from the connection connected to this decoder.
		msg, err := p.readMessage(reader)
		if err != nil {
			return
		}
		// Do something with the message.
		p.OnMessage(msg.From, msg)
	}
}

// Handle messages received.
func (p *Peer) OnMessage(from string, msg *Message) {
	// Sprintf because output is a channel.
	p.output <- fmt.Sprintf("Peer %s received %s (MsgID: %s) from Peer %s", p.id, msg.Type, msg.MsgID, from)
	// Pass message to demo synchronization.
	p.received <- *msg
	switch msg.Type {
	case helpers.PING_MESSAGE_TYPE:
		p.output <- fmt.Sprintf("Peer %s sending Pong (MsgID: %s) to Peer %s", p.id, msg.MsgID, from)
		p.Send(from, &Message{Type: helpers.PONG_MESSAGE_TYPE, MsgID: msg.MsgID, From: p.id})
		return // Do not flood ping messages.
	}
	p.addReceivedFloodMessage(msg)
	p.FloodNetwork(msg)
}

// Note this message as beeing received (should be used before calling FloodNetwork on a message).
func (p *Peer) addReceivedFloodMessage(msg *Message) {
	p.floodingLock.Lock()
	defer p.floodingLock.Unlock()
	var hist MessageHistory
	hist, exists := p.msgHistory[msg.MsgID]

	if !exists { // If the msg has never been received before, create a new history for it.
		hist = *NewMessageHistory(msg)
	}
	hist.receivedFrom = append(hist.receivedFrom, msg.From)
	p.msgHistory[msg.MsgID] = hist
}

// Flood a message across the network.
func (p *Peer) FloodNetwork(msg *Message) {
	p.floodingLock.Lock()
	defer p.floodingLock.Unlock()
	hist, _ := p.msgHistory[msg.MsgID]

	if hist.isSent { // If it did exist and was already sent, abort.
		return
	}
	// Set the message to be sent.
	hist.isSent = true
	p.msgHistory[msg.MsgID] = hist
	p.output <- fmt.Sprintf("Peer %s flooding %s (MsgID: %s)", p.id, msg.Type, msg.MsgID)
	// Change the sender, so that others are aware they received this message version from this peer.
	msg.From = p.id
	// Send to all peers it did not receive the message from (and also not itself).
	for peer, conn := range p.conns {
		if peer != p.id && !slices.Contains(hist.receivedFrom, peer) {
			p.writeMessage(conn, msg)
		}
	}
}

func (p *Peer) printOutput() {
	for msg := range p.output {
		fmt.Println(msg)
	}
}

// Connect to another peer.
func (p *Peer) Connect(addr string, port int) ([]string, error) {
	conn, err := net.Dial(helpers.PROTOCOL, addr+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	peers, err := p.prepareConnection(conn)
	if err != nil {
		return nil, err
	}
	return peers, nil
}

// Connect to a list of peers (may be known, in which case, it does notrhing).
func (p *Peer) joinNetwork(addr string, peers []string) {
	for _, peer := range peers {
		_, exists := p.conns[peer]
		if !exists {
			port, _ := strconv.Atoi(peer)
			p.Connect(addr, port)
		}
	}
}

// Send a message to another peer.
func (p *Peer) Send(to string, msg *Message) error {
	// Try to find the connection connected to the receiver of this message.
	p.lock.Lock()
	conn := p.conns[to]
	p.lock.Unlock()
	if conn == nil {
		return fmt.Errorf("Connection not found for receiver %s!", to)
	}
	// Send the message.
	return p.writeMessage(conn, msg)
}

// Marshal message and send it to the connection.
func (p *Peer) writeMessage(conn net.Conn, msg *Message) error {
	// Encode JSON to a buffer to compute length prefix.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(msg); err != nil {
		return err
	}
	data := buf.Bytes()
	header := make([]byte, helpers.MESSAGE_HEADER_SIZE)
	binary.BigEndian.PutUint32(header, uint32(len(data)))
	// Write header + payload atomically per connection.
	p.sendLock.Lock()
	defer p.sendLock.Unlock()
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

// Unmarshal message received from a connection.
func (p *Peer) readMessage(reader *bufio.Reader) (*Message, error) {
	// Read length prefix, then exact JSON payload.
	var header [helpers.MESSAGE_HEADER_SIZE]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header[:])
	if length == 0 {
		return nil, fmt.Errorf("invalid message length 0")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	var msg Message
	decoder := json.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
