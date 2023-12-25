package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/types"
	"github.com/Fantom-foundation/Substate/new_substate"
)

func NewMessage(message *new_substate.Message) *Message {
	m := &Message{
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

type Message struct {
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

	GasFeeCap *big.Int // missing in substate DB from Geth <= v1.10.3
	GasTipCap *big.Int // missing in substate DB from Geth <= v1.10.3
}

// ToSubstate transforms m from Message to new_substate.Message.
func (m Message) ToSubstate(getHashFunc func(codeHash common.Hash) ([]byte, error)) (*new_substate.Message, error) {
	sm := &new_substate.Message{
		Nonce:      m.Nonce,
		CheckNonce: m.CheckNonce,
		GasPrice:   m.GasPrice,
		Gas:        m.Gas,
		From:       m.From,
		To:         m.To,
		Value:      m.Value,
		Data:       m.Data,
		AccessList: m.AccessList,
		GasFeeCap:  m.GasFeeCap,
		GasTipCap:  m.GasTipCap,
	}

	// if receiver is nil, we have to extract the data from the DB using getHashFunc
	if sm.To == nil {
		var err error
		m.Data, err = getHashFunc(*m.InitCodeHash)
		if err != nil {
			return nil, err
		}
	}

	return sm, nil
}
