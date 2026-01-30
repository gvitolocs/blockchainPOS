package main

import (
	"net"
	"peer/helpers"
	"sync"
	"testing"
	"time"
)

func TestNewPeer(t *testing.T) {
	peer1 := NewPeer(42001)
	peer2 := NewPeer(42002)
	if peer1.listenport != 42001 {
		t.Errorf("Should be 42001")
	}
	if peer2.listenport != 42002 {
		t.Errorf("Should be 42002")
	}
}

func TestPeerStart(t *testing.T) {
	peer1 := NewPeer(42001)
	peer1.Start()
	conn1, _ := net.Dial(helpers.PROTOCOL, ":42001")
	defer conn1.Close()
	if conn1.RemoteAddr().String() != "127.0.0.1:42001" {
		t.Errorf("Expects connection to be 127.0.0.1:42001")
	}

	peer2 := NewPeer(42002)
	peer2.Start()
	conn2, _ := net.Dial(helpers.PROTOCOL, ":42002")
	defer conn2.Close()
	if conn2.RemoteAddr().String() != "127.0.0.1:42002" {
		t.Errorf("Expects connection to be 127.0.0.1:42002")
	}
}

func TestPeerConnect(t *testing.T) {
	peer1 := NewPeer(42001)
	peer1.Start()

	peer2 := NewPeer(42002)
	peer2.Start()

	var wg sync.WaitGroup
	wg.Go(func() {peer1.Connect("", peer2.listenport)})
	wg.Wait()
	peer1.Send("42002", &Message{From: "42001", Type: helpers.PING_MESSAGE_TYPE})
	time.Sleep(1 * time.Second) // TODO: Remove this line. We need some way to wait until pong-message is received.
	// TODO: We need something to test here.
}
