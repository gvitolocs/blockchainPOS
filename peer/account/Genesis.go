package account

import "encoding/json"

const BLOCKCHAIN_DEFAULT_HARDNESS = 10 // TODO: Change this.
const BLOCKCHAIN_SEED = 42
const NUMBER_OF_PRIME_ACCOUNTS = 10
const NUMBER_OF_SECONDARY_ACCOUNTS = 5

type GenesisMetaData struct {
	Hardness        int
	Seed            int
	InitialBalances map[string]int
}

func MakeGenesis() *Block {
	g := makeGenesisMetaData()
	data, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	strData := encode(data)
	return NewBlock(nil, 0, 0, "", strData)
}

func makeGenesisMetaData() *GenesisMetaData {
	g := new(GenesisMetaData)
	g.Hardness = BLOCKCHAIN_DEFAULT_HARDNESS
	g.Seed = BLOCKCHAIN_SEED
	createAccounts(g, NUMBER_OF_PRIME_ACCOUNTS, 10_000_000)
	createAccounts(g, NUMBER_OF_SECONDARY_ACCOUNTS, 0)
	return g
}

func createAccounts(genesis *GenesisMetaData, count int, initialBalance int) {
	for range count {
		account, err := NewAccount()
		if err != nil {
			panic(err)
		}
		genesis.InitialBalances[account.SafeEncode()] = initialBalance
	}
}

func (l *Ledger) InitializeFromGenesis(genesis *GenesisMetaData) {
	for account, amount := range genesis.InitialBalances {
		l.Accounts[account] = amount
	}
}
