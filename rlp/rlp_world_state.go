package rlp

import (
	"fmt"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
)

func NewWorldState(worldState substate.WorldState) WorldState {
	ws := WorldState{
		Addresses: []types.Address{},
		Accounts:  []*SubstateAccountRLP{},
	}

	for addr, acc := range worldState {
		ws.Addresses = append(ws.Addresses, addr)
		ws.Accounts = append(ws.Accounts, NewRLPAccount(acc))
	}

	return ws
}

type WorldState struct {
	Addresses []types.Address
	Accounts  []*SubstateAccountRLP
}

// ToSubstate transforms a from WorldState to substate.WorldState.
func (ws WorldState) ToSubstate(getHashFunc func(codeHash types.Hash) ([]byte, error)) (substate.WorldState, error) {
	sws := make(substate.WorldState)

	// iterate through addresses and assign it correctly to substate.WorldState
	// positions in WorldState match map assignment in substate.WorldState
	// that means that Address at first position matches SubstateAccountRLP at first position,
	// Address at second position matches SubstateAccountRLP at second position, and so on
	for i, addr := range ws.Addresses {
		acc := ws.Accounts[i]
		accCodeHash, err := getHashFunc(acc.CodeHash)
		if err != nil {
			return nil, fmt.Errorf("cannot get code hash; base code hash %s acc %s; %w", acc.CodeHash, addr, err)
		}
		sws[addr] = substate.NewAccount(acc.Nonce, acc.Balance, accCodeHash)
		for pos := range acc.Storage {
			sws[addr].Storage[acc.Storage[pos][0]] = acc.Storage[pos][1]
		}
	}

	return sws, nil
}
