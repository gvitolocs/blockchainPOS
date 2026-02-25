package main

import (
	"fmt"
	"net"
	"peer/helpers"
	"testing"
	"time"
)

func TestNewPeer(t *testing.T) {
	// Use free ports to avoid collisions on shared machines.
	port1 := getFreePort(t)
	port2 := getFreePort(t)
	peer1 := NewPeer(port1)
	peer2 := NewPeer(port2)
	if peer1.listenport != port1 {
		t.Errorf("Should be %d", port1)
	}
	if peer2.listenport != port2 {
		t.Errorf("Should be %d", port2)
	}
}

func TestPeerStart(t *testing.T) {
	// Start peers and verify listeners accept connections.
	port1 := getFreePort(t)
	port2 := getFreePort(t)
	peer1 := NewPeer(port1)
	peer1.Start()
	conn1, _ := net.Dial(helpers.PROTOCOL, fmt.Sprintf("127.0.0.1:%d", port1))
	// Defer ensures the connection is closed at test end.
	defer conn1.Close()
	if conn1.RemoteAddr().String() != fmt.Sprintf("127.0.0.1:%d", port1) {
		t.Errorf("Expects connection to be 127.0.0.1:%d", port1)
	}

	peer2 := NewPeer(port2)
	peer2.Start()
	conn2, _ := net.Dial(helpers.PROTOCOL, fmt.Sprintf("127.0.0.1:%d", port2))
	// Defer ensures the connection is closed at test end.
	defer conn2.Close()
	if conn2.RemoteAddr().String() != fmt.Sprintf("127.0.0.1:%d", port2) {
		t.Errorf("Expects connection to be 127.0.0.1:%d", port2)
	}
}

func TestPeerConnectAndSendMessages(t *testing.T) {
	// Connect two peers and verify k messages are received.
	port1 := getFreePort(t)
	port2 := getFreePort(t)
	peer1 := NewPeer(port1)
	peer1.Start()

	peer2 := NewPeer(port2)
	peer2.Start()

	if err := peer1.Connect("127.0.0.1", peer2.listenport); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !waitForConn(peer1, peer2.id, 2*time.Second) {
		t.Fatalf("Expected connection from %s in peer1", peer2.id)
	}
	if !waitForConn(peer2, peer1.id, 2*time.Second) {
		t.Fatalf("Expected connection from %s in peer2", peer1.id)
	}

	// Track MsgIDs to ensure we received all expected messages.
	wantIDs := map[string]bool{}
	k := 3
	// Cycle for sending k messages.
	for i := 0; i < k; i++ {
		msgID := fmt.Sprintf("msg-%d", i)
		wantIDs[msgID] = true
		if err := peer1.Send(peer2.id, &Message{
			Type:  helpers.PING_MESSAGE_TYPE,
			MsgID: msgID,
			From:  peer1.id,
		}); err != nil {
			t.Fatalf("Send failed: %v", err)
		}
	}

	// Wait for all messages with a timeout to avoid hanging tests.
	got := 0
	timeout := time.After(2 * time.Second)
	// Cycle for waiting until all messages arrive.
	for got < k {
		select {
		case msg := <-peer2.received:
			if msg.Type == helpers.PING_MESSAGE_TYPE && wantIDs[msg.MsgID] {
				got++
				delete(wantIDs, msg.MsgID)
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for messages, got %d/%d", got, k)
		}
	}
}

func getFreePort(t *testing.T) int {
	// Ask the OS for an available port.
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	// Defer ensures the listener is closed after we read the port.
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func waitForConn(peer *Peer, id string, timeout time.Duration) bool {
	// Poll until the connection appears or the timeout expires.
	deadline := time.Now().Add(timeout)
	// Cycle for polling the connection map.
	for time.Now().Before(deadline) {
		peer.lock.Lock()
		_, ok := peer.conns[id]
		peer.lock.Unlock()
		if ok {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
