package main

import (
	"fmt"
	"math/big"
	"os"
)

const WALLET_KEY_SIZE = 2048 / 8 // Key size in bytes.

/*
Create an  RSA sk/pk pair and store the sk in a file (given by filename)
encrypted under the password.
Returns the public key.
*/
func Generate(filename string, password string) string {
	// Create RSA key.
	n, e, d, _ := KeyGen(WALLET_KEY_SIZE * 8)
	sk := append(n.Bytes(), d.Bytes()...)
	pk := append(n.Bytes(), e.Bytes()...)
	// Hash the password.
	passwordHash := hashPassword(password)
	// Encrypt and save the secret key.
	EncryptToFile(passwordHash[:], sk, filename)
	return string(pk)
}

/*
Sign a message (msg) using the secret key stored in the file given by filename.
Returns the signature.
*/
func Sign(filename string, password string, msg []byte) []byte /*Signature type*/ {
	// Check file exists.
	if _, err := os.Stat(filename); err != nil {
		return nil
	}
	// Check password. TODO.

	// Get hold of the secret key stored in the file.
	passwordHash := hashPassword(password)
	sk, err := DecryptFromFile(passwordHash[:], filename)
	if err != nil {
		return nil
	}
	var n, d big.Int
	n.SetBytes(sk[:WALLET_KEY_SIZE])
	d.SetBytes(sk[WALLET_KEY_SIZE:])
	// Sign the message and return the signature.
	signature := RSASign(msg, &d, &n)
	return signature.Bytes()
}

// Wrapper around RSASign to handle public key and signature in string and byte format, respectively.
func VerifySignature(msg, signature []byte, pk string) bool {
	pkBytes := []byte(pk)
	var n, e, s big.Int
	n.SetBytes(pkBytes[:WALLET_KEY_SIZE])
	e.SetBytes(pkBytes[WALLET_KEY_SIZE:])
	s.SetBytes(signature)
	return RSAVerify(msg, &s, &e, &n)
}

// Convenience method to make a slow hashing method for password to slow adversary.
// TODO: Make this slow.
func hashPassword(password string) [32]byte {
	return applyHash([]byte(password))
}

func main() {
	password := "verygoodpassword(TM)"
	filename := "walletkey.sk"
	msg := []byte("An authentic message sent by DA1 Command HQ.")
	pk := Generate(filename, password)
	sign := Sign(filename, "password", msg)
	verify := VerifySignature(msg, sign, pk)

	fmt.Printf("Public key: %s\n\rSignature: %s\n\rVerification: %t", pk, string(sign), verify)
}
