package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"peer/helpers"
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
}

func NewPeer(listenport int) *Peer {
	peer := new(Peer)
	peer.id = strconv.Itoa(listenport)
	peer.listenport = listenport
	peer.conns = make(map[string]net.Conn)
	peer.output = make(chan string, 32)
	peer.received = make(chan Message, 32)
	return peer
}

func (p *Peer) Start() error {
	// Listen for connection.
	listener, err := net.Listen(helpers.PROTOCOL, ":"+strconv.Itoa(p.listenport))
	if err != nil {
		return err
	}
	// Goroutine to listen to any number of peers.
	go p.listenForPeers(listener)
	// Print logs from a single goroutine to avoid interleaving.
	go p.printOutput()
	return nil
}

func (p *Peer) listenForPeers(listener net.Listener) {
	// Defer ensures listener is closed when this goroutine returns.
	defer listener.Close()
	for {
		// Wait to establish connection.
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		// Prepare to marshal data for sending over network.
		p.prepareMarshalling(conn)
	}
}

func (p *Peer) prepareMarshalling(conn net.Conn) {
	reader := bufio.NewReader(conn)
	// Handshake: announce our id, then learn the peer id.
	p.lock.Lock() // Lock to ensure no messages get sent to this before it has had a chance to update decoders map.
	_ = p.writeMessage(conn, &Message{Type: helpers.CONNECT_MESSAGE_TYPE, From: p.id})
	// Wait for reply to establish their name.
	msg, err := p.readMessage(reader)
	if err != nil {
		p.lock.Unlock()
		_ = conn.Close()
		return
	}
	p.conns[msg.From] = conn
	p.lock.Unlock()
	go p.handleDecode(msg.From, conn, reader)
}

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

func (p *Peer) OnMessage(from string, msg *Message) {
	// Sprintf because output is a channel.
	p.output <- fmt.Sprintf("Peer %s received %s (MsgID: %s) from Peer %s", p.id, msg.Type, msg.MsgID, from)
	// Pass message to demo synchronization.
	p.received <- *msg
	switch msg.Type {
	case helpers.PING_MESSAGE_TYPE:
		p.output <- fmt.Sprintf("Peer %s sending Pong (MsgID: %s) to Peer %s", p.id, msg.MsgID, from)
		p.Send(from, &Message{Type: helpers.PONG_MESSAGE_TYPE, MsgID: msg.MsgID, From: p.id})
	}
}

func (p *Peer) printOutput() {
	for msg := range p.output {
		fmt.Println(msg)
	}
}

func (p *Peer) Connect(addr string, port int) error {
	conn, err := net.Dial(helpers.PROTOCOL, addr+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	p.prepareMarshalling(conn)
	return nil
}

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

func (p *Peer) writeMessage(conn net.Conn, msg *Message) error {
	// Encode JSON to a buffer to compute length prefix.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(msg); err != nil {
		return err
	}
	data := buf.Bytes()
	header := make([]byte, 4)
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

func (p *Peer) readMessage(reader *bufio.Reader) (*Message, error) {
	// Read length prefix, then exact JSON payload.
	var header [4]byte
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
