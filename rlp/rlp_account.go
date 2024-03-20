package rlp

import (
	"math/big"
	"sort"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types/common"
)

func NewRLPAccount(acc *substate.Account) *Account {
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
		return sortedKeys[i].Compare(sortedKeys[j]) < 0
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
