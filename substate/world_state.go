package substate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/holiman/uint256"
)

const (
	sizeOfAddress uint64 = 20
	sizeOfHash    uint64 = 32
	sizeOfNonce   uint64 = 8
)

func NewWorldState() WorldState {
	return make(map[types.Address]*Account)
}

type WorldState map[types.Address]*Account

// Add assigns new Account to an Address
func (ws WorldState) Add(addr types.Address, nonce uint64, balance *uint256.Int, code []byte) WorldState {
	ws[addr] = NewAccount(nonce, balance, code)
	return ws
}

// Merge y into ws. If values differs, values from y are saved.
func (ws WorldState) Merge(y WorldState) {
	for yAddr, yAcc := range y {
		if acc, found := ws[yAddr]; found {
			if acc.Equal(yAcc) {
				continue
			}

			// overwrite yAcc details in ws by y
			ws[yAddr].Nonce = yAcc.Nonce
			ws[yAddr].Balance = yAcc.Balance
			ws[yAddr].Code = make([]byte, len(yAcc.Code))
			copy(ws[yAddr].Code, yAcc.Code)
		} else {
			// create new yAcc details in a
			ws[yAddr] = NewAccount(yAcc.Nonce, yAcc.Balance, yAcc.Code)
		}
		// update storage by y
		for key, value := range yAcc.Storage {
			ws[yAddr].Storage[key] = value
		}
	}
}

// EstimateIncrementalSize returns estimated substate size increase after merge
func (ws WorldState) EstimateIncrementalSize(y WorldState) uint64 {
	var size uint64 = 0

	for yAddr, yAcc := range y {
		if acc, found := ws[yAddr]; found {
			// skip if no diff
			if acc.Equal(yAcc) {
				continue
			}
			// update storage by y
			for key := range yAcc.Storage {
				// only add new storage keys
				if _, found := ws[yAddr].Storage[key]; !found {
					size += sizeOfHash // add sizeof(types.Hash)
				}
			}
		} else {
			// add size of new accounts
			// address + nonce + balance + codehash
			size += sizeOfAddress + sizeOfNonce + uint64(len(yAcc.Balance.Bytes())) + sizeOfHash
			// storage slots * sizeof(types.Hash)
			size += uint64(len(yAcc.Storage)) * sizeOfHash
		}
	}
	return size
}

// Diff returns the difference set between two substate world state (z = ws\y).
// Note: Zero value and non-existing value are considered equal.
func (ws WorldState) Diff(y WorldState) WorldState {
	z := make(WorldState)
	for addr, acc := range ws {
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
					if yVal, found := y[addr].Storage[key]; (!found && value != types.Hash{}) || yVal != value {
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
func (ws WorldState) Equal(y WorldState) bool {
	if len(ws) != len(y) {
		return false
	}

	for key, val := range ws {
		yVal, exist := y[key]
		if !(exist && val.Equal(yVal)) {
			return false
		}
	}

	return true
}

func (ws WorldState) String() string {
	var builder strings.Builder

	for addr, acc := range ws {
		builder.WriteString(fmt.Sprintf("%s: %v", addr, acc.String()))
	}
	return builder.String()
}
