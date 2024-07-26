package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/types"
)

// legacySubstateRLP represents legacy RLP structure between before Berlin fork thus before berlinBlock
type legacySubstateRLP struct {
	InputAlloc  WorldState
	OutputAlloc WorldState
	Env         *legacyEnv
	Message     *legacyMessage
	Result      *Result
}

// toRLP transforms r into RLP format which is compatible with the currently used Geth fork.
func (r legacySubstateRLP) toRLP() *RLP {
	return &RLP{
		InputSubstate:  r.InputAlloc,
		OutputSubstate: r.OutputAlloc,
		Env:            r.Env.toEnv(),
		Message:        r.Message.toMessage(),
		Result:         r.Result,
	}
}

type legacyMessage struct {
	Nonce      uint64
	CheckNonce bool
	GasPrice   *big.Int
	Gas        uint64

	From  types.Address
	To    *types.Address `rlp:"nil"` // nil means contract creation
	Value *big.Int
	Data  []byte

	InitCodeHash *types.Hash `rlp:"nil"` // NOT nil for contract creation
}

// toMessage transforms m into RLP format which is compatible with the currently used Geth fork.
func (m legacyMessage) toMessage() *Message {
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
		AccessList:   nil, // access list was not present before berlin fork?

		// Same behavior as AccessListTx.gasFeeCap() and AccessListTx.gasTipCap()
		GasFeeCap: m.GasPrice,
		GasTipCap: m.GasPrice,
	}
}

type legacyEnv struct {
	Coinbase    types.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes [][2]types.Hash
}

// toEnv transforms e into RLP format which is compatible with the currently used Geth fork.
func (e legacyEnv) toEnv() *Env {
	return &Env{
		Coinbase:    e.Coinbase,
		Difficulty:  e.Difficulty,
		GasLimit:    e.GasLimit,
		Number:      e.Number,
		Timestamp:   e.Timestamp,
		BlockHashes: e.BlockHashes,
	}
}
