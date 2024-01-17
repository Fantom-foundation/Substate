package rlp

import (
	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/rlp"
	"github.com/Fantom-foundation/Substate/substate"
)

func NewRLP(substate *substate.Substate) *RLP {
	return &RLP{
		InputAlloc:  NewAlloc(substate.InputAlloc),
		OutputAlloc: NewAlloc(substate.OutputAlloc),
		Env:         NewEnv(substate.Env),
		Message:     NewMessage(substate.Message),
		Result:      NewResult(substate.Result),
	}
}

type RLP struct {
	InputAlloc  Alloc
	OutputAlloc Alloc
	Env         *Env
	Message     *Message
	Result      *Result
}

// Decode decodes val into RLP and returns it.
func Decode(val []byte, block uint64) (*RLP, error) {
	var (
		substateRLP RLP
		err         error
	)

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
	if err == nil {
		return nil, err
	}

	return legacy.toLondon(), nil
}

// ToSubstate transforms every attribute of r from RLP to substate.Substate.
func (r RLP) ToSubstate(getHashFunc func(codeHash common.Hash) ([]byte, error), block uint64, tx int) (*substate.Substate, error) {
	msg, err := r.Message.ToSubstate(getHashFunc)
	if err != nil {
		return nil, err
	}

	return &substate.Substate{
		InputAlloc:  r.InputAlloc.ToSubstate(),
		OutputAlloc: r.OutputAlloc.ToSubstate(),
		Env:         r.Env.ToSubstate(),
		Message:     msg,
		Result:      r.Result.ToSubstate(),
		Block:       block,
		Transaction: tx,
	}, nil
}
