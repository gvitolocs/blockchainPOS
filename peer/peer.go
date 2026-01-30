package main

import (
	"encoding/json"
	"fmt"
	"net"
	"peer/helpers"
	"strconv"
	"sync"
)

type Peer struct {
	listenport int
	id string
	encoders map[string]*json.Encoder
	lock sync.Mutex
}

func NewPeer(listenport int) *Peer {
	peer := new(Peer)
	peer.id = strconv.Itoa(listenport)
	peer.listenport = listenport
	peer.encoders = make(map[string]*json.Encoder)
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
	return nil
}

func (p *Peer) listenForPeers(listener net.Listener) {
	defer listener.Close()
	for {
		// Wait to establish connection.
		conn, err := listener.Accept() // TODO: Someone should ensure connections are closed.
		if err != nil {
			panic(err)
		}
		// Prepare to marshal data for sending over network.
		p.prepareMarshalling(conn)
	}
}

func (p *Peer) prepareMarshalling(conn net.Conn) {
	// Create en/de-coders for sending messages.
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	// Send a message to notify connection of the name of this.
	p.lock.Lock()
	encoder.Encode(Message{Type: helpers.CONNECT_MESSAGE_TYPE, From: p.id})
	// Wait for reply to establish their name.
	var msg Message
	decoder.Decode(&msg)
	p.encoders[msg.From] = encoder
	p.lock.Unlock()
	go p.handleDecode(decoder)
}

func (p *Peer) handleDecode(decoder *json.Decoder) {
	var msg Message
	for {
		// Wait until receving a message from the connection connected to this decoder.
		err := decoder.Decode(&msg)
		if err != nil {
			return
		}
		// Do something with the message.
		p.OnMessage(msg.From, &msg)
	}
}

func (p *Peer) OnMessage(from string, msg *Message) {
	switch msg.Type {
	case helpers.PING_MESSAGE_TYPE:
		p.Send(from, &Message{Type: helpers.PONG_MESSAGE_TYPE, From: p.id})
	}
}

func (p* Peer) Connect(addr string, port int) error {
	conn, err := net.Dial(helpers.PROTOCOL, addr + ":" + strconv.Itoa(port)) // TODO: Someone should ensure connections are closed.
	if err != nil {
		return err
	}
	p.prepareMarshalling(conn)
	return err
}

func (p *Peer) Send(to string, msg *Message) error {
	// Try to find the encoder connected to the receiver of this message.
	encoder := p.encoders[to]
	if encoder == nil {
		return fmt.Errorf("Encoder not found for receiver %s!", to)
	}
	fmt.Println("HEL", msg)
	// Send the message.
	encoder.Encode(msg)
	return nil
}