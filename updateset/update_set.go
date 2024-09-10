package updateset

import (
	"errors"

	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/syndtr/goleveldb/leveldb"
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
	DeletedAccounts []types.Address
}

func (s UpdateSet) ToWorldStateRLP() rlp.WorldState {
	a := rlp.WorldState{
		Addresses: []types.Address{},
		Accounts:  []*rlp.SubstateAccountRLP{},
	}

	for addr, acc := range s.WorldState {
		a.Addresses = append(a.Addresses, addr)
		a.Accounts = append(a.Accounts, rlp.NewRLPAccount(acc))
	}

	return a
}

func NewUpdateSetRLP(updateSet *UpdateSet, deletedAccounts []types.Address) UpdateSetRLP {
	return UpdateSetRLP{
		WorldState:      updateSet.ToWorldStateRLP(),
		DeletedAccounts: deletedAccounts,
	}
}

// UpdateSetRLP represents the DB structure of UpdateSet.
type UpdateSetRLP struct {
	WorldState      rlp.WorldState
	DeletedAccounts []types.Address
}

func (up UpdateSetRLP) ToWorldState(getCodeFunc func(codeHash types.Hash) ([]byte, error), block uint64) (*UpdateSet, error) {
	worldState := make(substate.WorldState)

	for i, addr := range up.WorldState.Addresses {
		worldStateAcc := up.WorldState.Accounts[i]

		code, err := getCodeFunc(worldStateAcc.CodeHash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, err
		}

		acc := substate.Account{
			Nonce:   worldStateAcc.Nonce,
			Balance: types.BigIntToUint256(worldStateAcc.Balance),
			Storage: make(map[types.Hash]types.Hash),
			Code:    code,
		}

		for j := range worldStateAcc.Storage {
			acc.Storage[worldStateAcc.Storage[j][0]] = worldStateAcc.Storage[j][1]
		}
		worldState[addr] = &acc
	}

	return NewUpdateSet(worldState, block), nil
}

func (x *UpdateSet) Equal(y *UpdateSet) bool {
	if x == y {
		return true
	}
	if !x.WorldState.Equal(y.WorldState) {
		return false
	}

	if x.Block != y.Block {
		return false
	}

	for i, val := range x.DeletedAccounts {
		if val != y.DeletedAccounts[i] {
			return false
		}
	}
	return true
}
