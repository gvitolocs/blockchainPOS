package main

import (
	"fmt"
	"peer/helpers"
	"sync"
)

func main() {
	//Starting peers
	peer1 := NewPeer(42003)
	err := peer1.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42003, err))
	}
	fmt.Printf("Peer %s started and listening on port %d\n", peer1.id, 42003)

	peer2 := NewPeer(42004)
	err2 := peer2.Start()
	if err2 != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42004, err2))
	}
	fmt.Printf("Peer %s started and listening on port %d\n", peer2.id, 42004)

	peer3 := NewPeer(42005)
	err3 := peer3.Start()
	if err3 != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42005, err3))
	}
	fmt.Printf("Peer %s started and listening on port %d\n", peer3.id, 42005)

	//Connecting peers
	fmt.Printf("Peer %s connecting to Peer %s\n", peer1.id, peer2.id)
	peer1.Connect("localhost", 42004)

	fmt.Printf("Peer %s connecting to Peer %s\n", peer1.id, peer3.id)
	peer1.Connect("localhost", 42005)

	fmt.Printf("Peer %s connecting to Peer %s\n", peer2.id, peer3.id)
	peer2.Connect("localhost", 42005)

	// Demonstrate ping/pong message exchange between peers
	var wg sync.WaitGroup
	wg.Add(2)

	// Peer 1 sends Ping to Peer 2
	msgID1 := "msg-001"
	fmt.Printf("Peer %s sending Ping (MsgID: %s) to Peer %s\n", peer1.id, msgID1, peer2.id)
	peer1.Send(peer2.id, &Message{
		Type:  helpers.PING_MESSAGE_TYPE,
		MsgID: msgID1,
		From:  peer1.id,
	})

	// Peer 3 sends Ping to Peer 1
	msgID2 := "msg-002"
	fmt.Printf("Peer %s sending Ping (MsgID: %s) to Peer %s\n", peer3.id, msgID2, peer1.id)
	peer3.Send(peer1.id, &Message{
		Type:  helpers.PING_MESSAGE_TYPE,
		MsgID: msgID2,
		From:  peer3.id,
	})

	go waitForPong(peer1, msgID1, &wg)
	go waitForPong(peer3, msgID2, &wg)
	wg.Wait()
	fmt.Println("\n Demo complete")
}

func waitForPong(peer *Peer, msgID string, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range peer.received {
		if msg.Type == helpers.PONG_MESSAGE_TYPE && msg.MsgID == msgID {
			return
		}
	}
}
