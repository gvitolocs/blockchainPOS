package main

import (
	"fmt"
	"net"
	"peer/helpers"
	"testing"
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
	peer1 := NewPeer(42003)
	peer1.Start()

	peer2 := NewPeer(42004)
	peer2.Start()

	peer1.Connect("", peer2.listenport)
	_, has_key_42002 := peer1.encoders[peer2.id]
	if !has_key_42002 {
		t.Errorf("Expected key %s, but found %s.", peer2.id, fmt.Sprint(peer1.encoders))
	}
	_, has_key_42001 := peer2.encoders[peer1.id]
	if !has_key_42001 {
		t.Errorf("Expected key %s, but found %s.", peer1.id, fmt.Sprint(peer2.encoders))
	}
}
