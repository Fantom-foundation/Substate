package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/types"
)

const (
	berlinBlock = 37_455_223
	londonBlock = 37_534_833
)

// IsLondonFork returns true if block is part of the london fork block range
func IsLondonFork(block uint64) bool {
	return block >= londonBlock
}

// IsBerlinFork returns true if block is part of the berlin fork block range
func IsBerlinFork(block uint64) bool {
	return block >= berlinBlock && block < londonBlock
}

// legacyRLP represents legacy RLP structure between before Berlin fork thus before berlinBlock
type legacyRLP struct {
	InputAlloc  Alloc
	OutputAlloc Alloc
	Env         *legacyEnv
	Message     *legacyMessage
	Result      *Result
}

func (r legacyRLP) toLondon() *RLP {
	return &RLP{
		InputAlloc:  r.InputAlloc,
		OutputAlloc: r.OutputAlloc,
		Env:         r.Env.toLondon(),
		Message:     r.Message.toLondon(),
		Result:      r.Result,
	}
}

type legacyMessage struct {
	Nonce      uint64
	CheckNonce bool
	GasPrice   *big.Int
	Gas        uint64

	From  common.Address
	To    *common.Address `rlp:"nil"` // nil means contract creation
	Value *big.Int
	Data  []byte

	InitCodeHash *common.Hash `rlp:"nil"` // NOT nil for contract creation
}

func (m legacyMessage) toLondon() *Message {
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
	Coinbase    common.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes [][2]common.Hash
}

func (e legacyEnv) toLondon() *Env {
	return &Env{
		Coinbase:    e.Coinbase,
		Difficulty:  e.Difficulty,
		GasLimit:    e.GasLimit,
		Number:      e.Number,
		Timestamp:   e.Timestamp,
		BlockHashes: e.BlockHashes,
	}
}

// berlinRLP represents legacy RLP structure between Berlin and London fork starting at berlinBlock ending at londonBlock
type berlinRLP struct {
	InputAlloc  Alloc
	OutputAlloc Alloc
	Env         *legacyEnv
	Message     *berlinMessage
	Result      *Result
}

func (r berlinRLP) toLondon() *RLP {
	return &RLP{
		InputAlloc:  r.InputAlloc,
		OutputAlloc: r.OutputAlloc,
		Env:         r.Env.toLondon(),
		Message:     r.Message.toLondon(),
		Result:      r.Result,
	}

}

type berlinMessage struct {
	Nonce      uint64
	CheckNonce bool
	GasPrice   *big.Int
	Gas        uint64

	From  common.Address
	To    *common.Address `rlp:"nil"` // nil means contract creation
	Value *big.Int
	Data  []byte

	InitCodeHash *common.Hash `rlp:"nil"` // NOT nil for contract creation

	AccessList types.AccessList // missing in substate DB from Geth v1.9.x
}

func (m berlinMessage) toLondon() *Message {
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
