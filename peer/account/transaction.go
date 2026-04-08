package account

import (
	"encoding/json"
	"math/big"
	"peer/dissycrypto"
)

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

type SignedTransaction struct {
	Transact  *Transaction
	Signature []byte
}

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

func NewSignedTransaction(ID string, From *Account, To string, Amount int) *SignedTransaction {
	s := new(SignedTransaction)
	tx := new(Transaction)
	tx.ID = ID
	tx.From = From.Encode()
	tx.To = To
	tx.Amount = Amount
	s.Transact = tx
	s.signTransaction(From.d, From.n)
	return s
}

func (s *SignedTransaction) signTransaction(d, n *big.Int) {
	data, err := json.Marshal(s.Transact)
	if err != nil {
		return
	}
	s.Signature = dissycrypto.RSASign(data, d, n).Bytes()
}

func (s *SignedTransaction) Verify(pk string) bool {
	data, err := json.Marshal(s.Transact)
	if err != nil {
		return false
	}
	return dissycrypto.VerifySignature(data, s.Signature, pk)
}
