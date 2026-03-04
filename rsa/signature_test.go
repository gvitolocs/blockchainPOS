package main

import (
	"crypto/sha256"
	"math/big"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	msg := "A message to sign."
	n, e, d, _ := KeyGen(256)
	sign := RSASign([]byte(msg), d, n)
	if !RSAVerify([]byte(msg), sign, e, n) {
		t.Errorf("Message not verified, but shoud have been!")
	}
	badMsg := "An aletered message with different content."
	if RSAVerify([]byte(badMsg), sign, e, n) {
		t.Errorf("Bad message was verified!")
	}
	if RSAVerify([]byte{0}, big.NewInt(0), e, n) { // Just showing that (m, sign) pairs that are always valid in normal RSA-signature are not valid here.
		t.Errorf("Bad message was verified!")
	}
	if RSAVerify([]byte{1}, big.NewInt(1), e, n) { // Just showing that (m, sign) pairs that are always valid in normal RSA-signature are not valid here.
		t.Errorf("Bad message was verified!")
	}
}

func BenchmarkHashSpeed(b *testing.B) {
	msg := make([]byte, 0, 10_000)
	sha256.Sum256(msg)
}

/* Result:
goos: windows
goarch: amd64
pkg: dissycrypto
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkHashSpeed-8   	1000000000	         0.0000065 ns/op	       0 B/op	       0 allocs/op
*/

func BenchmarkSignatureOnHash(b *testing.B) {
	msg := make([]byte, 0, 10_000)
	n, _, d, _ := KeyGen(2000)
	RSASign(msg, d, n)
}

/* Result:
goos: windows
goarch: amd64
pkg: dissycrypto
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkSignatureOnHash-8   	     536	   1970926 ns/op	   52031 B/op	     133 allocs/op
*/
