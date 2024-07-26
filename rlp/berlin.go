package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/types"
)

// berlinRLP represents legacy RLP structure between Berlin and London fork starting at berlinBlock ending at londonBlock
type berlinRLP struct {
	InputAlloc  WorldState
	OutputAlloc WorldState
	Env         *legacyEnv
	Message     *berlinMessage
	Result      *Result
}

// toRLP transforms r into RLP format which is compatible with the currently used Geth fork.
func (r berlinRLP) toRLP() *RLP {
	return &RLP{
		InputSubstate:  r.InputAlloc,
		OutputSubstate: r.OutputAlloc,
		Env:            r.Env.toEnv(),
		Message:        r.Message.toMessage(),
		Result:         r.Result,
	}

}

type berlinMessage struct {
	Nonce      uint64
	CheckNonce bool
	GasPrice   *big.Int
	Gas        uint64

	From  types.Address
	To    *types.Address `rlp:"nil"` // nil means contract creation
	Value *big.Int
	Data  []byte

	InitCodeHash *types.Hash `rlp:"nil"` // NOT nil for contract creation

	AccessList types.AccessList // missing in substate DB from Geth v1.9.x
}

// toMessage transforms m into RLP format which is compatible with the currently used Geth fork.
func (m berlinMessage) toMessage() *Message {
	return &Message{
		Nonce:        m.Nonce,
		CheckNonce:   m.CheckNonce,
		GasPrice:     m.GasPrice,
		Gas:          m.Gas,
		From:         m.From,
		To:           m.To,
		Value:        new(big.Int).Set(m.Value),
		Data:         m.Data,
		InitCodeHash: m.InitCodeHash,
		AccessList:   m.AccessList,

		// Same behavior as AccessListTx.gasFeeCap() and AccessListTx.gasTipCap()
		GasFeeCap: m.GasPrice,
		GasTipCap: m.GasPrice,
	}
}
