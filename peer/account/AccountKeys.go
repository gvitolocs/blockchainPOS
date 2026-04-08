package account

import (
	"math/big"
	"peer/dissycrypto"
)

type Account struct {
	n *big.Int
	e *big.Int
	d *big.Int
}

func NewUser() (*Account, error) {
	n, e, d, err := dissycrypto.KeyGen(dissycrypto.WALLET_KEY_SIZE * 8)
	if err != nil {
		return nil, err
	}
	a := new(Account)
	a.n = n
	a.e = e
	a.d = d
	return a, nil
}

func (a *Account) Encode() string {
	return string(append(a.n.Bytes(), a.e.Bytes()...)) //base64.StdEncoding.EncodeToString(append(a.n.Bytes(), a.e.Bytes()...))
}
