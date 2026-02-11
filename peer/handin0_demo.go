package main

import (
	"fmt"
	"peer/helpers"
	"sync"
	"time"
)

func main() {

	fmt.Println("\nStarting 3 peers")
	//Starting peers
	peer1 := NewPeer(42003)
	err := peer1.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42003, err))
	}
	fmt.Printf("Peer 1 started and listening on port %s\n", peer1.id)

	peer2 := NewPeer(42004)
	err2 := peer2.Start()
	if err2 != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42004, err))
	}
	fmt.Printf("Peer 2 started and listening on port %s\n", peer2.id)

	peer3 := NewPeer(42005)
	err3 := peer3.Start()
	if err3 != nil {
		panic(fmt.Sprintf("Failed to start peer on port %d: %v", 42005, err))
	}
	fmt.Printf("Peer 3 started and listening on port %s\n", peer3.id)

	var wg sync.WaitGroup

	// Start logger goroutines
	wg.Add(3)
	go logger(peer1.output, &wg)
	go logger(peer2.output, &wg)
	go logger(peer3.output, &wg)

	//Connecting peers
	fmt.Println("\nConnecting peers")

	fmt.Printf("Peer 1 on port %s connecting to Peer 2 on port %s\n", peer1.id, peer2.id)
	peer1.Connect("localhost", 42004)

	fmt.Printf("Peer 1 on port%s connecting to Peer 3 on port %s\n", peer1.id, peer3.id)
	peer1.Connect("localhost", 42005)

	fmt.Printf("Peer 2 on port %s connecting to Peer 3 on port %s\n", peer2.id, peer3.id)
	peer2.Connect("localhost", 42005)

	// Demonstrate ping/pong message exchange between peers

	fmt.Println("\nDemonstrating Ping/Pong message exchange")

	// Peer 1 sends Ping to Peer 2
	msgID1 := "msg-001"
	fmt.Printf("Peer 1 on port %s sending Ping (MsgID: %s) to Peer 2 on port %s\n", peer1.id, msgID1, peer2.id)
	peer1.Send(peer2.id, &Message{
		Type:  helpers.PING_MESSAGE_TYPE,
		MsgID: msgID1,
		From:  peer1.id,
	})

	// Peer 3 sends Ping to Peer 1
	msgID2 := "msg-002"
	fmt.Printf("Peer 3 on port %s sending Ping (MsgID: %s) to Peer 1 on port %s\n", peer3.id, msgID2, peer1.id)
	peer3.Send(peer1.id, &Message{
		Type:  helpers.PING_MESSAGE_TYPE,
		MsgID: msgID2,
		From:  peer3.id,
	})

	time.Sleep(500 * time.Millisecond)

	close(peer1.output)
	close(peer2.output)
	close(peer3.output)
	wg.Wait()

	fmt.Println("\n Demo complete")

}

func logger(output chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range output {
		fmt.Println(msg)
	}
}
