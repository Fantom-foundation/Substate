package substate

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Fantom-foundation/Substate/geth/common"
)

const (
	sizeOfAddress uint64 = 20
	sizeOfHash    uint64 = 32
	sizeOfNonce   uint64 = 8
)

func NewAlloc() Alloc {
	return make(map[common.Address]*Account)
}

type Alloc map[common.Address]*Account

// Add assigns new Account to an Address
func (a Alloc) Add(addr common.Address, nonce uint64, balance *big.Int, code []byte) Alloc {
	a[addr] = NewAccount(nonce, balance, code)
	return a
}

// Merge y into a. If values differs, values from y are saved.
func (a Alloc) Merge(y Alloc) {
	for yAddr, yAcc := range y {
		if acc, found := a[yAddr]; found {
			if acc.Equal(yAcc) {
				continue
			}

			// overwrite yAcc details in a by y
			a[yAddr].Nonce = yAcc.Nonce
			a[yAddr].Balance = new(big.Int).Set(yAcc.Balance)
			a[yAddr].Code = make([]byte, len(yAcc.Code))
			copy(a[yAddr].Code, yAcc.Code)
		} else {
			// create new yAcc details in a
			a[yAddr] = NewAccount(yAcc.Nonce, yAcc.Balance, yAcc.Code)
		}
		// update storage by y
		for key, value := range yAcc.Storage {
			a[yAddr].Storage[key] = value
		}
	}
}

// EstimateIncrementalSize returns estimated substate size increase after merge
func (a Alloc) EstimateIncrementalSize(y Alloc) uint64 {
	var size uint64 = 0

	for yAddr, yAcc := range y {
		if acc, found := a[yAddr]; found {
			// skip if no diff
			if acc.Equal(yAcc) {
				continue
			}
			// update storage by y
			for key, _ := range yAcc.Storage {
				// only add new storage keys
				if _, found := a[yAddr].Storage[key]; !found {
					size += sizeOfHash // add sizeof(common.Hash)
				}
			}
		} else {
			// add size of new accounts
			// address + nonce + balance + codehash
			size += sizeOfAddress + sizeOfNonce + uint64(len(yAcc.Balance.Bytes())) + sizeOfHash
			// storage slots * sizeof(common.Hash)
			size += uint64(len(yAcc.Storage)) * sizeOfHash
		}
	}
	return size
}

// Diff returns the difference set between two substate alloc (z = a\y).
// Note: Zero value and non-existing value are considered equal.
func (a Alloc) Diff(y Alloc) Alloc {
	z := make(Alloc)
	for addr, acc := range a {
		if yAcc, found := y[addr]; !found {
			z[addr] = acc.Copy()
		} else {
			if yAcc.Equal(acc) {
				continue
			} else {
				// check nonce, balance and code
				equal := acc.Nonce == yAcc.Nonce &&
					acc.Balance.Cmp(yAcc.Balance) == 0 &&
					bytes.Equal(acc.Code, yAcc.Code)
				if !equal {
					z[addr] = NewAccount(acc.Nonce, acc.Balance, acc.Code)
				}

				// check storage
				for key, value := range acc.Storage {
					if yVal, found := y[addr].Storage[key]; (!found && value != common.Hash{}) || yVal != value {
						// initialize if not exists.
						if _, found := z[addr]; !found {
							z[addr] = NewAccount(acc.Nonce, acc.Balance, acc.Code)
						}
						z[addr].Storage[key] = value
					}
				}
			}
		}
	}
	return z
}

// Equal returns true if a is y or if values of a are equal to values of y.
// Otherwise, a and y are not equal hence false is returned.
func (a Alloc) Equal(y Alloc) bool {
	if len(a) != len(y) {
		return false
	}

	for key, val := range a {
		yVal, exist := y[key]
		if !(exist && val.Equal(yVal)) {
			return false
		}
	}

	return true
}

func (a Alloc) String() string {
	var builder strings.Builder

	for addr, acc := range a {
		builder.WriteString(fmt.Sprintf("%v: %v", addr.Hex(), acc.String()))
	}
	return builder.String()
}
