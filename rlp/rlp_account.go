package rlp

import (
	"math/big"
	"sort"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/new_substate"
)

func NewRLPAccount(acc *new_substate.Account) *Account {
	a := &Account{
		Nonce:    acc.Nonce,
		Balance:  new(big.Int).Set(acc.Balance),
		CodeHash: acc.CodeHash(),
		Storage:  [][2]common.Hash{},
	}

	var sortedKeys []common.Hash
	for key := range acc.Storage {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].Big().Cmp(sortedKeys[j].Big()) < 0
	})

	for _, key := range sortedKeys {
		value := acc.Storage[key]
		a.Storage = append(a.Storage, [2]common.Hash{key, value})
	}

	return a
}

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash common.Hash
	Storage  [][2]common.Hash
}
