package rlp

import (
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/rlp"
)

func NewRLP(substate *substate.Substate) *RLP {
	return &RLP{
		PreState:  NewWorldState(substate.PreState),
		PostState: NewWorldState(substate.PostState),
		Env:       NewEnv(substate.Env),
		Message:   NewMessage(substate.Message),
		Result:    NewResult(substate.Result),
	}
}

type RLP struct {
	PreState  WorldState
	PostState WorldState
	Env       *Env
	Message   *Message
	Result    *Result
}

// Decode decodes val into RLP and returns it.
func Decode(val []byte, block uint64) (*RLP, error) {
	var (
		substateRLP RLP
		err         error
	)
	// todo decode does not work
	err = rlp.DecodeBytes(val, &substateRLP)
	if err == nil {
		return &substateRLP, nil
	}

	var berlin berlinRLP
	err = rlp.DecodeBytes(val, &berlin)
	if err == nil {
		return berlin.toLondon(), nil
	}

	var legacy legacyRLP
	err = rlp.DecodeBytes(val, &legacy)
	if err != nil {
		return nil, err
	}

	return legacy.toLondon(), nil
}

// ToSubstate transforms every attribute of r from RLP to substate.Substate.
func (r RLP) ToSubstate(getHashFunc func(codeHash types.Hash) ([]byte, error), block uint64, tx int) (*substate.Substate, error) {
	msg, err := r.Message.ToSubstate(getHashFunc)
	if err != nil {
		return nil, err
	}

	return &substate.Substate{
		PreState:    r.PreState.ToSubstate(),
		PostState:   r.PostState.ToSubstate(),
		Env:         r.Env.ToSubstate(),
		Message:     msg,
		Result:      r.Result.ToSubstate(),
		Block:       block,
		Transaction: tx,
	}, nil
}
