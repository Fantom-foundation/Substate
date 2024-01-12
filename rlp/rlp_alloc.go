package rlp

import (
	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/substate"
)

func NewAlloc(alloc substate.Alloc) Alloc {
	a := Alloc{
		Addresses: []common.Address{},
		Accounts:  []*Account{},
	}

	for addr, acc := range alloc {
		a.Addresses = append(a.Addresses, addr)
		a.Accounts = append(a.Accounts, NewRLPAccount(acc))
	}

	return a
}

type Alloc struct {
	Addresses []common.Address
	Accounts  []*Account
}

// ToSubstate transforms a from Alloc to substate.Alloc.
func (a Alloc) ToSubstate() substate.Alloc {
	sa := make(substate.Alloc)

	// iterate through addresses and assign it correctly to substate.Alloc
	// positions in Alloc match map assignment in substate.Alloc
	// that means that Address at first position matches Account at first position,
	// Address at second position matches Account at second position, and so on
	for i, addr := range a.Addresses {
		acc := a.Accounts[i]
		sa[addr] = substate.NewAccount(acc.Nonce, acc.Balance, acc.CodeHash.Bytes())
		for pos := range acc.Storage {
			sa[addr].Storage[acc.Storage[pos][0]] = acc.Storage[pos][1]
		}
	}

	return sa
}
