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
	id         string
	encoders   map[string]*json.Encoder
	output     chan string
	lock       sync.Mutex
}

func NewPeer(listenport int) *Peer {
	peer := new(Peer)
	peer.id = strconv.Itoa(listenport)
	peer.listenport = listenport
	peer.encoders = make(map[string]*json.Encoder)
	peer.output = make(chan string, 100)
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
	p.lock.Lock() // Lock to ensure no messages get sent to this before it has had a chance to update decoders map.
	encoder.Encode(Message{Type: helpers.CONNECT_MESSAGE_TYPE, From: p.id})
	// Wait for reply to establish their name.
	var msg Message
	decoder.Decode(&msg)
	p.encoders[msg.From] = encoder
	// fmt.Println(p.encoders, p.id)
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
	//fmt.Println("on", msg)
	switch msg.Type {
	case helpers.PING_MESSAGE_TYPE:
		logMsg := fmt.Sprintf("Peer %s received Ping (MsgID: %s) from Peer %s", p.id, msg.MsgID, from)
		p.output <- logMsg

		p.Send(from, &Message{Type: helpers.PONG_MESSAGE_TYPE, MsgID: msg.MsgID, From: p.id})

		sendMsg := fmt.Sprintf("Peer %s sent Pong (MsgID: %s) to Peer %s", p.id, msg.MsgID, from)
		p.output <- sendMsg

	case helpers.PONG_MESSAGE_TYPE:
		logMsg := fmt.Sprintf("Peer %s received Pong (MsgID: %s) from Peer %s", p.id, msg.MsgID, from)
		p.output <- logMsg
	}
}

func (p *Peer) Connect(addr string, port int) error {
	conn, err := net.Dial(helpers.PROTOCOL, addr+":"+strconv.Itoa(port)) // TODO: Someone should ensure connections are closed.
	if err != nil {
		return err
	}
	p.prepareMarshalling(conn)
	return nil
}

func (p *Peer) Send(to string, msg *Message) error {
	// Try to find the encoder connected to the receiver of this message.
	encoder := p.encoders[to]
	if encoder == nil {
		return fmt.Errorf("Encoder not found for receiver %s!", to)
	}
	// Send the message.
	encoder.Encode(msg)
	return nil
}
