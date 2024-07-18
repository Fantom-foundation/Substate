package rlp

import (
	"errors"
	"math/big"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewMessage(message *substate.Message) *Message {
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

	From  types.Address
	To    *types.Address `rlp:"nil"` // nil means contract creation
	Value *big.Int
	Data  []byte

	InitCodeHash *types.Hash `rlp:"nil"` // NOT nil for contract creation

	AccessList types.AccessList // missing in substate DB from Geth v1.9.x

	GasFeeCap *big.Int // missing in substate DB from Geth <= v1.10.3
	GasTipCap *big.Int // missing in substate DB from Geth <= v1.10.3

	BlobGasFeeCap *big.Int     // missing in substate DB from Geth before Cancun
	BlobHashes    []types.Hash // missing in substate DB from Geth before Cancun
}

// ToSubstate transforms m from Message to substate.Message.
func (m Message) ToSubstate(getHashFunc func(codeHash types.Hash) ([]byte, error)) (*substate.Message, error) {
	sm := &substate.Message{
		Nonce:         m.Nonce,
		CheckNonce:    m.CheckNonce,
		GasPrice:      m.GasPrice,
		Gas:           m.Gas,
		From:          m.From,
		To:            m.To,
		Value:         m.Value,
		Data:          m.Data,
		AccessList:    m.AccessList,
		GasFeeCap:     m.GasFeeCap,
		GasTipCap:     m.GasTipCap,
		BlobGasFeeCap: m.BlobGasFeeCap,
		BlobHashes:    m.BlobHashes,
	}

	// if receiver is nil, we have to extract the data from the DB using getHashFunc
	if sm.To == nil {
		var err error
		sm.Data, err = getHashFunc(*m.InitCodeHash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, err
		}
	}

	return sm, nil
}
