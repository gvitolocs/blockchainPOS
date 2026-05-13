package account

import (
	"crypto/sha256"
	"encoding/json"
	"peer/dissycrypto"
	"strconv"
	"sync"
)

const BLOCK_SIZE = 16 // The maximum number of transactions in a block.
const BLOCK_TYPE = "block"

type Block struct {
	Type            string // Tuple type: The text "block".
	VerificationKey string // Verification key for block winner.
	Slot            int    // Slot number.
	Draw            int    // The draw used to win the lottery.
	MetaData        string // The block metadata. For Genesis, this is settings. Otherwise, transactions.
	ParentHash      string // The hash of the parent block.
	Signature       string // A signature on the block for authentication from the winner.
}

func NewBlock(winner *Account, slot int, draw int, parentHash string, metaData string) *Block {
	b := new(Block)
	b.Type = BLOCK_TYPE
	b.Slot = slot
	b.Draw = draw
	b.MetaData = metaData
	b.ParentHash = parentHash
	if winner != nil { // Useful for Genesis. Otherwise, this will make the block invalid.
		b.VerificationKey = winner.SafeEncode()
		b.Signature = b.Sign(winner)
	}
	return b
}

func (b *Block) marshalForSignature() []byte {
	msg, err := json.Marshal([]string{b.Type, b.VerificationKey, strconv.Itoa(b.Slot),
		strconv.Itoa(b.Draw), b.ParentHash})
	if err != nil {
		panic(err)
	}
	return msg
}

func (b *Block) MarshalBlock() []byte {
	msg := b.marshalForSignature()
	signature, err := decode(b.Signature)
	if err != nil {
		panic(err)
	}
	msg = append(msg, signature...)
	return msg
}

func (b *Block) GetBlockHash() [32]byte {
	return sha256.Sum256(b.MarshalBlock())
}

func (b *Block) Sign(signer *Account) string {
	signature := dissycrypto.RSASign(b.marshalForSignature(), signer.d, signer.n)
	return encode(signature.Bytes())
}

type Blockchain struct {
	Blocks []Block
	lock   sync.Mutex
}

func NewBlockchain() *Blockchain {
	b := new(Blockchain)
	meta, err := json.Marshal(makeGenesisMetaData())
	if err != nil {
		panic(err)
	}
	b.Blocks = append(b.Blocks, Block{
		BLOCK_TYPE,
		"",
		0,
		0,
		encode(meta),
		"",
		"",
	})
	return b
}

// Get all transactions in the current longest branch (except those in the <rollback> latest blocks).
func (b *Blockchain) GetCurrentTransactions(rollback int) []string {
	b.lock.Lock()
	defer b.lock.Unlock()

	// Get transactions in the current longest branch.
	var transactions []string
	for _, block := range b.Blocks {
		transactions = append(transactions, block.MetaData) // TODO: Wrong but useful start.
	}
	return transactions
}

// Given a list of transactions in Total Order, create a ledger.
func LedgerFromTransactions(transactions []string) *Ledger {
	ledger := MakeLedger()
	var genesis GenesisMetaData
	json.Unmarshal([]byte(transactions[0]), &genesis)
	ledger.InitializeFromGenesis(&genesis)
	if len(transactions) == 1 {
		return ledger
	}
	// Apply transactions.
	for _, txString := range transactions[1:] {
		var tx SignedTransaction
		json.Unmarshal([]byte(txString), &tx)
		ledger.Transaction(&tx)
	}
	return ledger
}
