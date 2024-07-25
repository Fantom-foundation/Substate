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
		londonMessage: newLondonMessage(message),
		BlobGasFeeCap: message.BlobGasFeeCap,
		BlobHashes:    message.BlobHashes,
	}

	return m
}

type Message struct {
	londonMessage

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
