// handin.go — Demo for Exercise 4.6 (Hand-in 2): simple P2P ledger.
package main

import (
	"fmt"
	"maps"
	"math/rand"
	"peer/account"
	"peer/helpers"
	"sync"
	"time"
)

// main is the entry point for the hand-in demo (traccia: "program handin.go").
// Run: go run .
func main() {
	runHandin()
}

// runHandin starts peers, floods transactions, and checks ledger convergence.
func runHandin() {
	n := 5
	tau := 5
	const NUM_ACCOUNTS = 6
	accounts := [NUM_ACCOUNTS]*account.Account{}
	for i := range NUM_ACCOUNTS {
		acc, err := account.NewUser()
		if err != nil {
			return
		}
		accounts[i] = acc
	}
	basePort := 43000
	peers := make([]*Peer, n)

	// 1) Start first peer with no existing network (it starts its own).
	peers[0] = NewPeer(basePort)
	peers[0].StartWithConnection("localhost", -1)

	// 2) Start the rest and point them to the first peer so they join the same network.
	for i := 1; i < n; i++ {
		peers[i] = NewPeer(basePort + i)
		peers[i].StartWithConnection("localhost", basePort)
	}

	// Give the network time to form (Join messages and connections).
	time.Sleep(500 * time.Millisecond)

	totalTx := n * tau
	// 3) Start counters before sending so no early transaction event is missed.
	// Each peer expects totalTx events because FloodTransaction now emits one local event too.
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		p := peers[i]
		go func(peer *Peer) {
			count := 0
			for count < totalTx {
				msg := <-peer.received
				if msg.Type == helpers.TRANSACTION_MESSAGE_TYPE {
					count++
				}
			}
			done <- struct{}{}
		}(p)
	}

	// 4) Each peer sends τ transactions; all use the same 5 accounts.
	var wgSend sync.WaitGroup
	for i := 0; i < n; i++ {
		p := peers[i]
		wgSend.Add(1)
		go func(peer *Peer, peerIdx int) {
			defer wgSend.Done()
			for j := 0; j < tau; j++ {
				// Simple tx: rotate between accounts so we use the same 5.
				from := accounts[(peerIdx+j)%len(accounts)]
				to := accounts[(peerIdx+j+1)%len(accounts)]
				tx := account.NewSignedTransaction(
					fmt.Sprintf("tx-%s-%d-%d", peer.id, peerIdx, j),
					from,
					to.Encode(),
					rand.Intn(10))
				peer.FloodTransaction(tx)
			}
		}(p, i)
	}
	wgSend.Wait()
	fmt.Printf("Submitted %d total transactions across %d peers\n", totalTx, n)

	// 5) Wait until every peer observed totalTx transaction events.
	// This is a delivery/completion check before final ledger comparison.
	for i := 0; i < n; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			fmt.Println("Timeout waiting for transactions")
			return
		}
	}
	fmt.Println("All transactions delivered at all peers")

	// Give handlers time to apply all transactions before we compare ledgers.
	time.Sleep(1 * time.Second)

	// 6) Check that all peers have the same ledger (compare first peer with the rest).
	ref := peers[0].ledger.CopyAccounts()
	// Print all ledgers once at the end: useful for TA verification, low noise.
	fmt.Println("Final ledgers:")
	fmt.Printf("  Peer %s: %v\n", peers[0].id, ref)
	for i := 1; i < n; i++ {
		other := peers[i].ledger.CopyAccounts()
		fmt.Printf("  Peer %s: %v\n", peers[i].id, other)
		if !maps.Equal(ref, other) {
			fmt.Printf("Ledger mismatch: peer %s differs from peer %s\n", peers[i].id, peers[0].id)
			return
		}
	}
	fmt.Println("All ledgers identical. Demo complete.")
}
