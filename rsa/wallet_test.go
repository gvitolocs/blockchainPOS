package main

import "testing"

func TestGenerate(t *testing.T) {
	TEST_WALLET_KEY_FILENAME := "./testfiles/walletkey-test1.sk"
	password := "verygoodpassword(TM)"
	pk := Generate(TEST_WALLET_KEY_FILENAME, password)
	pkBytes := []byte(pk)

	passwordHash := hashPassword(password)
	sk, _ := DecryptFromFile(passwordHash[:], TEST_WALLET_KEY_FILENAME)
	if !compareSlices(pkBytes[:WALLET_KEY_SIZE], sk[:WALLET_KEY_SIZE]) {
		t.Errorf("sk and pk do not share n!")
	}
}

func TestGenerateAndSign(t *testing.T) {
	TEST_WALLET_KEY_FILENAME := "./testfiles/walletkey-test2.sk"
	CORRECT_PASSWORD := "verygoodpassword(TM)"
	pk := Generate(TEST_WALLET_KEY_FILENAME, CORRECT_PASSWORD)

	msg := []byte("An authentic message.")
	wrongSignature := Sign(TEST_WALLET_KEY_FILENAME, "wrong password", msg)
	correctSignature := Sign(TEST_WALLET_KEY_FILENAME, CORRECT_PASSWORD, msg)
	if VerifySignature(msg, wrongSignature, pk) {
		t.Errorf("Wrong signature passed test!")
	}
	if !VerifySignature(msg, correctSignature, pk) {
		t.Errorf("Correct signature failed test!")
	}
}

func compareSlices(v1, v2 []byte) bool {
	if len(v1) != len(v2) {
		return false
	}
	for idx := range len(v1) {
		if v1[idx] != v2[idx] {
			return false
		}
	}
	return true
}
