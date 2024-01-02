package update_set

import (
	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/new_substate"
	"github.com/Fantom-foundation/Substate/rlp"
)

func NewUpdateSet(alloc new_substate.Alloc, block uint64) *UpdateSet {
	return &UpdateSet{
		Alloc: alloc,
		Block: block,
	}
}

// UpdateSet represents the new_substate.Account allocation for the block.
type UpdateSet struct {
	Alloc           new_substate.Alloc
	Block           uint64
	DeletedAccounts []common.Address
}

func (s UpdateSet) ToAlloc() rlp.Alloc {
	a := rlp.Alloc{
		Addresses: []common.Address{},
		Accounts:  []*rlp.Account{},
	}

	for addr, acc := range s.Alloc {
		a.Addresses = append(a.Addresses, addr)
		a.Accounts = append(a.Accounts, rlp.NewRLPAccount(acc))
	}

	return a
}

func NewUpdateSetRLP(updateSet *UpdateSet, deletedAccounts []common.Address) UpdateSetRLP {
	return UpdateSetRLP{
		Alloc:           updateSet.ToAlloc(),
		DeletedAccounts: deletedAccounts,
	}
}

// UpdateSetRLP represents the DB structure of UpdateSet.
type UpdateSetRLP struct {
	Alloc           rlp.Alloc
	DeletedAccounts []common.Address
}

func (up UpdateSetRLP) ToSubstateAlloc(getCodeFunc func(codeHash common.Hash) ([]byte, error), block uint64) (*UpdateSet, error) {
	alloc := make(new_substate.Alloc)

	for i, addr := range up.Alloc.Addresses {
		allocAcc := up.Alloc.Accounts[i]

		code, err := getCodeFunc(allocAcc.CodeHash)
		if err != nil {
			return nil, err
		}

		acc := new_substate.Account{
			Nonce:   allocAcc.Nonce,
			Balance: allocAcc.Balance,
			Storage: make(map[common.Hash]common.Hash),
			Code:    code,
		}

		for j := range allocAcc.Storage {
			acc.Storage[up.Alloc.Accounts[j].Storage[j][0]] = up.Alloc.Accounts[j].Storage[j][1]
		}
		alloc[addr] = &acc
	}

	return NewUpdateSet(alloc, block), nil
}
