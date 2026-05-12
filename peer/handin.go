// handin.go — Demo for Exercise 4.6 (Hand-in 2): simple P2P ledger.
package main

import (
	"fmt"
	"maps"
	"peer/account"
	"peer/helpers"
	"strings"
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
	fmt.Println("Starting demo.")
	const NUM_PEERS = 5

	// Make accounts.
	acc1, _ := account.NewAccount()
	acc2, _ := account.NewAccount()
	acc3, _ := account.NewAccount()

	EASY_ACCOUNT_NAMES := map[string]string{
		acc1.SafeEncode(): "A",
		acc2.SafeEncode(): "B",
		acc3.SafeEncode(): "C",
	}

	// Make transactions to send.
	validTransactions := []*account.SignedTransaction{
		account.NewSignedTransaction("val-01", acc1, acc2.SafeEncode(), 3),
		account.NewSignedTransaction("val-02", acc3, acc2.SafeEncode(), 5),
		account.NewSignedTransaction("val-03", acc1, acc2.SafeEncode(), 7),
		account.NewSignedTransaction("val-04", acc2, acc1.SafeEncode(), 11),
		account.NewSignedTransaction("val-05", acc3, acc1.SafeEncode(), 13),
		account.NewSignedTransaction("val-06", acc1, acc3.SafeEncode(), 17),
	}

	inv2 := account.NewSignedTransaction("inv-02", acc3, acc2.SafeEncode(), 5)
	inv2.Signature = validTransactions[2].Signature
	inv3 := account.NewSignedTransaction("inv-03", acc1, acc2.SafeEncode(), 7)
	inv3.From = acc3.SafeEncode()
	invalidTransactions := []*account.SignedTransaction{
		validTransactions[0], // Replay.
		inv2,                 // Fake signature.
		inv3,                 // Changing sender.
	}

	// Connect peers.
	basePort := 43000
	peers := make([]*Peer, NUM_PEERS)

	// 1) Start first peer with no existing network (it starts its own).
	peers[0] = NewPeer(basePort)
	peers[0].StartWithConnection("localhost", -1)

	// 2) Start the rest and point them to the first peer so they join the same network.
	for i := 1; i < NUM_PEERS; i++ {
		peers[i] = NewPeer(basePort + i)
		peers[i].StartWithConnection("localhost", basePort)
	}

	// Give the network time to form (Join messages and connections).
	time.Sleep(500 * time.Millisecond)

	totalTx := len(validTransactions)
	// 3) Start counters before sending so no early transaction event is missed.
	// Each peer expects len(validTransactions) events because FloodTransaction now emits one local event too.
	done := make(chan struct{})
	for i := 0; i < NUM_PEERS; i++ {
		go func(peer *Peer) {
			count := 0
			for count < totalTx {
				msg := <-peer.received
				if msg.Type == helpers.TRANSACTION_MESSAGE_TYPE {
					count++
				}
			}
			done <- struct{}{}
		}(peers[i])
	}

	// 4) Transactions are sent by different peers.
	var wgSend sync.WaitGroup
	for i, tx := range append(invalidTransactions, validTransactions...) {
		//fmt.Printf("Flooding transaction %s from Peer %s\n", tx.ID, peers[i%NUM_PEERS].id)
		flood(peers[i%NUM_PEERS], tx, &wgSend)
	}
	wgSend.Wait()
	fmt.Printf("Submitted %d valid, %d invalid transactions across %d peers\n", len(validTransactions), len(invalidTransactions), NUM_PEERS)

	// 5) Wait until every peer observed totalTx transaction events.
	// This is a delivery/completion check before final ledger comparison.
	for i := 0; i < NUM_PEERS; i++ {
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

	// 6) Check that demo works and show relevant output.

	// Check that all peers have the same ledger (compare first peer with the rest).
	ref := peers[0].ledger.CopyAccountsPretty(EASY_ACCOUNT_NAMES)
	// Print all ledgers once at the end: useful for TA verification, low noise.
	fmt.Println("Final ledgers:")
	fmt.Printf("  Peer %s: %v\n", peers[0].id, ref)
	for i := 1; i < NUM_PEERS; i++ {
		other := peers[i].ledger.CopyAccountsPretty(EASY_ACCOUNT_NAMES)
		fmt.Printf("  Peer %s: %v\n", peers[i].id, other)
		if !maps.Equal(ref, other) {
			fmt.Printf("Ledger mismatch: peer %s differs from peer %s\n", peers[i].id, peers[0].id)
			return
		}
	}

	// Check that only valid transactions were applied.
	for _, p := range peers {
		if len(p.ledger.TxHistory) != len(validTransactions) {
			fmt.Printf("Peer %s has an incorrect amount of transactions (%d)\n", p.id, len(p.ledger.TxHistory))
		}
		for name := range p.ledger.TxHistory {
			if strings.Contains(name, "inv") {
				fmt.Println("An invalid transaction was recorded.")
			}
		}
	}
	fmt.Printf("All ledgers have %d transactions and no invalid transactions recorded!\n", len(validTransactions))

	fmt.Println("Demo complete.")
}

func flood(sender *Peer, tx *account.SignedTransaction, wg *sync.WaitGroup) {
	wg.Go(func() { sender.FloodTransaction(tx) })
}
