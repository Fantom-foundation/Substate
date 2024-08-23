package rlp

import (
	"errors"
	"math/big"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewMessage(sm *substate.Message) *Message {
	mess := &Message{
		Nonce:         sm.Nonce,
		CheckNonce:    sm.CheckNonce,
		GasPrice:      sm.GasPrice,
		Gas:           sm.Gas,
		From:          sm.From,
		To:            sm.To,
		Value:         new(big.Int).Set(sm.Value),
		Data:          sm.Data,
		AccessList:    sm.AccessList,
		GasFeeCap:     sm.GasFeeCap,
		GasTipCap:     sm.GasTipCap,
		BlobGasFeeCap: sm.BlobGasFeeCap,
		BlobHashes:    sm.BlobHashes,
	}

	if mess.To == nil {
		// put contract creation init code into codeDB
		dataHash := sm.DataHash()
		mess.InitCodeHash = &dataHash
		mess.Data = nil
	}

	return mess
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
