package rlp

import (
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
)

func NewWorldState(worldState substate.WorldState) WorldState {
	ws := WorldState{
		Addresses: []types.Address{},
		Accounts:  []*Account{},
	}

	for addr, acc := range worldState {
		ws.Addresses = append(ws.Addresses, addr)
		ws.Accounts = append(ws.Accounts, NewRLPAccount(acc))
	}

	return ws
}

type WorldState struct {
	Addresses []types.Address
	Accounts  []*Account
}

// ToSubstate transforms a from WorldState to substate.WorldState.
func (ws WorldState) ToSubstate() substate.WorldState {
	sws := make(substate.WorldState)

	// iterate through addresses and assign it correctly to substate.WorldState
	// positions in WorldState match map assignment in substate.WorldState
	// that means that Address at first position matches Account at first position,
	// Address at second position matches Account at second position, and so on
	for i, addr := range ws.Addresses {
		acc := ws.Accounts[i]
		sws[addr] = substate.NewAccount(acc.Nonce, acc.Balance, acc.CodeHash[:])
		for pos := range acc.Storage {
			sws[addr].Storage[acc.Storage[pos][0]] = acc.Storage[pos][1]
		}
	}

	return sws
}
