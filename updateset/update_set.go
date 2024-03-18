package updateset

import (
	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/Fantom-foundation/Substate/substate"
)

func NewUpdateSet(alloc substate.WorldState, block uint64) *UpdateSet {
	return &UpdateSet{
		WorldState: alloc,
		Block:      block,
	}
}

// UpdateSet represents the substate.Account world state for the block.
type UpdateSet struct {
	WorldState      substate.WorldState
	Block           uint64
	DeletedAccounts []common.Address
}

func (s UpdateSet) ToWorldStateRLP() rlp.WorldState {
	a := rlp.WorldState{
		Addresses: []common.Address{},
		Accounts:  []*rlp.Account{},
	}

	for addr, acc := range s.WorldState {
		a.Addresses = append(a.Addresses, addr)
		a.Accounts = append(a.Accounts, rlp.NewRLPAccount(acc))
	}

	return a
}

func NewUpdateSetRLP(updateSet *UpdateSet, deletedAccounts []common.Address) UpdateSetRLP {
	return UpdateSetRLP{
		WorldState:      updateSet.ToWorldStateRLP(),
		DeletedAccounts: deletedAccounts,
	}
}

// UpdateSetRLP represents the DB structure of UpdateSet.
type UpdateSetRLP struct {
	WorldState      rlp.WorldState
	DeletedAccounts []common.Address
}

func (up UpdateSetRLP) ToWorldState(getCodeFunc func(codeHash common.Hash) ([]byte, error), block uint64) (*UpdateSet, error) {
	worldState := make(substate.WorldState)

	for i, addr := range up.WorldState.Addresses {
		worldStateAcc := up.WorldState.Accounts[i]

		code, err := getCodeFunc(worldStateAcc.CodeHash)
		if err != nil {
			return nil, err
		}

		acc := substate.Account{
			Nonce:   worldStateAcc.Nonce,
			Balance: worldStateAcc.Balance,
			Storage: make(map[common.Hash]common.Hash),
			Code:    code,
		}

		for j := range worldStateAcc.Storage {
			acc.Storage[up.WorldState.Accounts[j].Storage[j][0]] = up.WorldState.Accounts[j].Storage[j][1]
		}
		worldState[addr] = &acc
	}

	return NewUpdateSet(worldState, block), nil
}
