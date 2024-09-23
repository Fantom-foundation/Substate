package substate

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/hash"
	"github.com/holiman/uint256"
)

// Account holds any information about account used in a transaction.
type Account struct {
	Nonce   uint64
	Balance *big.Int
	Storage map[types.Hash]types.Hash
	Code    []byte
}

func NewAccount(nonce uint64, balance *big.Int, code []byte) *Account {
	return &Account{
		Nonce:   nonce,
		Balance: balance,
		Storage: make(map[types.Hash]types.Hash),
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
		yValue, exists := y.Storage[aKey]
		if !(exists && aVal == yValue) {
			return false
		}
	}

	return true
}

// Copy returns a hard copy of a
func (a *Account) Copy() *Account {
	accCopy := NewAccount(a.Nonce, a.Balance, a.Code)

	for key, value := range a.Storage {
		accCopy.Storage[key] = value
	}

	return accCopy
}

// CodeHash returns hashed code
func (a *Account) CodeHash() types.Hash {
	return hash.Keccak256Hash(a.Code)
}

func (a *Account) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Nonce: %v\nBalance: %v\nCode: %v\nStorage:", a.Nonce, a.Balance.String(), string(a.Code)))

	for key, val := range a.Storage {
		builder.WriteString(fmt.Sprintf("%s: %s\n", key, val))
	}

	return builder.String()
}
