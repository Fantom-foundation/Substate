package new_substate

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/crypto"
)

// Account holds any information about account used in a transaction.
type Account struct {
	Nonce   uint64
	Balance *big.Int
	Storage map[common.Hash]common.Hash
	Code    []byte
}

func NewAccount(nonce uint64, balance *big.Int, code []byte) *Account {
	return &Account{
		Nonce:   nonce,
		Balance: balance,
		Storage: make(map[common.Hash]common.Hash),
		Code:    code,
	}
}

// Equal returns true if a is y or if values of a are equal to values of y.
// Otherwise, a and y are not equal hence false is returned.
func (a *Account) Equal(y *Account) bool {
	if a == y {
		return true
	}

	if (a == nil || y == nil) && a != y {
		return false
	}

	// check values
	equal := a.Nonce == y.Nonce &&
		a.Balance.Cmp(y.Balance) == 0 &&
		bytes.Equal(a.Code, y.Code) &&
		len(a.Storage) == len(y.Storage)
	if !equal {
		return false
	}

	for aKey, aVal := range a.Storage {
		yValue, exist := y.Storage[aKey]
		if !(exist && aVal == yValue) {
			return false
		}
	}

	return true
}

// Copy returns a hard copy of a
func (a *Account) Copy() *Account {
	cpy := NewAccount(a.Nonce, a.Balance, a.Code)

	for key, value := range a.Storage {
		cpy.Storage[key] = value
	}

	return cpy
}

// CodeHash returns hashed code
func (a *Account) CodeHash() common.Hash {
	return crypto.Keccak256Hash(a.Code)
}

func (a *Account) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Nonce: %v\nBalance: %v\nCode: %v\nStorage:", a.Nonce, a.Balance.String(), string(a.Code)))

	for key, val := range a.Storage {
		builder.WriteString(fmt.Sprintf("%v: %v\n", key.Hex(), val.Hex()))
	}

	return builder.String()
}
