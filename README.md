# Hand-in 10 - Exercise 16.2 (Static Proof-of-Stake Blockchain)

This repository contains the implementation of **Exercise 16.2** from the ADNO course material, focused on a simple but complete **tree-based blockchain with static Proof-of-Stake (PoS)**.

The project starts from the previous peer-to-peer flooding ledger and extends it into a PoS blockchain system with:

- slot-based block production,
- longest-chain total ordering,
- block and transaction validation,
- miner rewards,
- throughput measurement,
- rollback observation.

## Context from the exercise

Following the exercise specification, the implementation uses:

- a genesis configuration with **10 initial accounts**, each with **1,000,000 AU**,
- fixed `Hardness` and `Seed` for deterministic lottery checks,
- one-second slots,
- transaction validity rules (signature, positive amount, no overdraft),
- block rewards:
  - `+10 AU` per accepted block,
  - `+1 AU` per transaction included in that block.

## What was implemented

### 1) Blockchain core (`peer/account`)

- Tree-based blockchain data structure.
- Deterministic PoS lottery draw and win check.
- Full block validation pipeline:
  - parent existence,
  - block signature verification,
  - lottery correctness,
  - sequential transaction validity against parent state.
- Longest-chain traversal and ledger reconstruction from genesis.
- Rollback support by ignoring the last `k` blocks of the best branch.

### 2) Peer integration (`peer/peer.go`)

- PoS transaction propagation (`PoSTransaction`) into mempool.
- Slot mining (`MineOneSlot`) and block flooding.
- Block handling with re-validation and deterministic ledger rebuild from best chain.
- Deduplication and safer connection cleanup for concurrent network behavior.

### 3) Demo and validation (`peer/handin.go`, tests)

- End-to-end demo with 10 peers.
- Valid and invalid transaction scenarios.
- Throughput printing (valid tx/s on best chain).
- Rollback comparison (`rollback=0` vs `rollback=1`).
- Added/updated tests for blockchain behavior and convergence checks.

## Main engineering steps addressed

During implementation and debugging, the main issues addressed were:

- Non-deterministic genesis serialization (fixed to keep chain roots aligned across peers).
- Runtime blocking due to message/channel pressure in network handling.
- Longest-path and metadata traversal bugs in partially implemented blockchain code.
- Consistency issues between demo behavior and test expectations.

The final result is stable for `go test ./...` and for the handin demo run.

## Run the demo

```bash
cd peer
go run .
```

Expected output includes:

- final ledgers for each peer,
- mined block count,
- throughput metric,
- rollback comparison,
- completion message.

## Run tests

```bash
cd peer
go test ./...
```

## Report

- PDF: `reports/reportHI10.pdf`
- Source: `reports/reportHI10.md`
