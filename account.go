package substate

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Account holds any information about account used in a transaction.
type Account struct {
	Nonce   uint64
	Balance *big.Int
	Storage map[common.Hash]common.Hash
	Code    []byte
}

func NewAccount(nonce uint64, balance *big.Int, storage map[common.Hash]common.Hash, code []byte) *Account {
	return &Account{
		Nonce:   nonce,
		Balance: balance,
		Storage: storage,
		Code:    code,
	}
}
