package main

import (
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
	if RSAVerify([]byte{0}, big.NewInt(0), e, n) {
		t.Errorf("Bad message was verified!")
	}
	if RSAVerify([]byte{1}, big.NewInt(1), e, n) {
		t.Errorf("Bad message was verified!")
	}
}
