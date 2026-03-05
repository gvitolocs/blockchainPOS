---
title: "Validation and Measurements Report"
author: "Hand-in 3"
date: "\\today"
documentclass: article
geometry:
  - margin=2.5cm
  - a4paper
fontsize: 11pt
numbersections: true
---

# Validation and Measurements

## Verify signature generation and verification

We can generate and verify a signature on a message.

- **Positive test:** Sign a message, verify with the same message -> passes.
- **Negative test:** Modify the message and verify again -> verification fails.

See `rsa/signature_test.go` - `TestVerifySignature`.

**Run:** `cd rsa && go test -run TestVerifySignature -v`

---

## Hashing speed (bits per second)

- **Message size:** 10 KB = 10,000 bytes = 80,000 bits  
- **Measured:** 4,200 ns per 10 KB hash  
- **Throughput:** 80,000 / (4,200 x 10^-9^) = 1.95 x 10^10^ bits/s = **19 Gbit/s**

**Run:** `cd rsa && go test -bench=BenchmarkHashSpeed -benchmem`

---

## RSA signature generation time (2000-bit key)

Signing a hash value (256 bits) with a 2000-bit RSA key:

- **Time per signature:** 1,970,926 ns = **1.97 ms**

**Run:** `cd rsa && go test -bench=BenchmarkSignatureOnHash -benchmem`

---

## RSA throughput vs hashing

**If RSA were used directly on the whole message:** at most 2000 bits per RSA operation. With ~1.97 ms per op:

- **RSA-direct throughput:** 2000 / 0.00197 = 1.14 x 10^6^ bits/s = **1.14 Mbit/s**

**Hashing (from item 2):** ~19 Gbit/s.

Hashing is far faster. With **hash-then-sign** we sign the 256-bit hash once per message; the cost of hashing is negligible compared to RSA. **Hashing greatly improves efficiency.**

---

# How to run tests and the project

**Run all RSA tests** (key size, encrypt/decrypt, signatures, AES, hybrid):

```bash
cd rsa
go test
```

**Run signature verification only:**

```bash
cd rsa
go test -run TestVerifySignature -v
```

**Run hashing speed benchmark** (10 KB, ~4,200 ns -> ~19 Gbit/s):

```bash
cd rsa
go test -bench=BenchmarkHashSpeed -benchmem
```

**Run RSA signature benchmark** (2000-bit key, ~1.97 ms per signature):

```bash
cd rsa
go test -bench=BenchmarkSignatureOnHash -benchmem
```

**Run peer (P2P flooding + ledger) tests:**

```bash
cd peer
go test
```

**Run the peer demo** (5 peers, 5 transactions each, ledger convergence check):

```bash
cd peer
go run .
```
