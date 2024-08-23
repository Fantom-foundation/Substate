package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
)

func NewLondonRLP(substate *substate.Substate) *londonRLP {
	return &londonRLP{
		InputSubstate:  NewWorldState(substate.InputSubstate),
		OutputSubstate: NewWorldState(substate.OutputSubstate),
		Env:            newLondonEnv(substate.Env),
		Message:        newLondonMessage(substate.Message),
		Result:         NewResult(substate.Result),
	}
}

// londonRLP represents RLP structure after londonBlock and before cancun fork.
type londonRLP struct {
	InputSubstate  WorldState
	OutputSubstate WorldState
	Env            londonEnv
	Message        londonMessage
	Result         *Result
}

// toRLP transforms r into RLP format which is compatible with the currently used Geth fork.
func (r londonRLP) toRLP() *RLP {
	return &RLP{
		InputSubstate:  r.InputSubstate,
		OutputSubstate: r.OutputSubstate,
		Env:            r.Env.toEnv(),
		Message:        r.Message.toMessage(),
		Result:         r.Result,
	}
}

func newLondonEnv(env *substate.Env) londonEnv {
	e := londonEnv{
		Coinbase:    env.Coinbase,
		Difficulty:  env.Difficulty,
		GasLimit:    env.GasLimit,
		Number:      env.Number,
		Timestamp:   env.Timestamp,
		BlockHashes: createBlockHashes(env.BlockHashes),
	}

	e.BaseFee = nil
	if env.BaseFee != nil {
		baseFeeHash := types.BigToHash(env.BaseFee)
		e.BaseFee = &baseFeeHash
	}

	return e
}

func createBlockHashes(m map[uint64]types.Hash) (blockHashes [][2]types.Hash) {
	var sortedNum64 []uint64
	for num64 := range m {
		sortedNum64 = append(sortedNum64, num64)
	}

	for _, num64 := range sortedNum64 {
		num := types.BigToHash(new(big.Int).SetUint64(num64))
		blockHash := m[num64]
		pair := [2]types.Hash{num, blockHash}
		blockHashes = append(blockHashes, pair)
	}
	return blockHashes
}

type londonEnv struct {
	Coinbase    types.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes [][2]types.Hash

	BaseFee *types.Hash `rlp:"nil"` // missing in substate DB from Geth <= v1.10.3
}

// toEnv transforms m into RLP format which is compatible with the currently used Geth fork.
func (e londonEnv) toEnv() *Env {
	return &Env{
		Coinbase:    e.Coinbase,
		Difficulty:  e.Difficulty,
		GasLimit:    e.GasLimit,
		Number:      e.Number,
		Timestamp:   e.Timestamp,
		BlockHashes: e.BlockHashes,
		BaseFee:     e.BaseFee,
	}
}

func newLondonMessage(message *substate.Message) londonMessage {
	m := londonMessage{
		Nonce:        message.Nonce,
		CheckNonce:   message.CheckNonce,
		GasPrice:     message.GasPrice,
		Gas:          message.Gas,
		From:         message.From,
		To:           message.To,
		Value:        new(big.Int).Set(message.Value),
		Data:         message.Data,
		InitCodeHash: nil,
		AccessList:   message.AccessList,
		GasFeeCap:    message.GasFeeCap,
		GasTipCap:    message.GasTipCap,
	}

	if m.To == nil {
		// put contract creation init code into codeDB
		dataHash := message.DataHash()
		m.InitCodeHash = &dataHash
		m.Data = nil
	}

	return m
}

type londonMessage struct {
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

	GasFeeCap *big.Int // missing in substate DB from Geth <= v1.10.3
	GasTipCap *big.Int // missing in substate DB from Geth <= v1.10.3
}

// toMessage transforms m into RLP format which is compatible with the currently used Geth fork.
func (m londonMessage) toMessage() *Message {
	return &Message{
		Nonce:        m.Nonce,
		CheckNonce:   m.CheckNonce,
		GasPrice:     m.GasPrice,
		Gas:          m.Gas,
		From:         m.From,
		To:           m.To,
		Value:        m.Value,
		Data:         m.Data,
		InitCodeHash: m.InitCodeHash,
		AccessList:   m.AccessList,
		GasFeeCap:    m.GasFeeCap,
		GasTipCap:    m.GasFeeCap,
	}
}
