package rlp

import (
	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/new_substate"
)

func NewAlloc(alloc new_substate.Alloc) Alloc {
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

// ToSubstate transforms a from Alloc to new_substate.Alloc.
func (a Alloc) ToSubstate() new_substate.Alloc {
	sa := make(new_substate.Alloc)

	// iterate through addresses and assign it correctly to new_substate.Alloc
	// positions in Alloc match map assignment in new_substate.Alloc
	// that means that Address at first position matches Account at first position,
	// Address at second position matches Account at second position, and so on
	for i, addr := range a.Addresses {
		acc := a.Accounts[i]
		sa[addr] = new_substate.NewAccount(acc.Nonce, acc.Balance, acc.CodeHash.Bytes())
		for pos := range acc.Storage {
			sa[addr].Storage[acc.Storage[pos][0]] = acc.Storage[pos][1]
		}
	}

	return sa
}
