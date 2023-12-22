package substate

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/rlp"
	"github.com/Fantom-foundation/Substate/geth/types"
)

const (
	berlinBlock = 37_455_223
	londonBlock = 37_534_833
)

// ----------------------------------------------------------------------------
//                                     RLP
// ----------------------------------------------------------------------------

// todo use new substate
func NewRLP(substate *Substate) *RLP {
	panic("not yet implemented")
}

type RLP struct {
	InputAlloc  RLPAlloc
	OutputAlloc RLPAlloc
	Env         *RLPEnv
	Message     *RLPMessage
	Result      *RLPResult
}

type RLPAlloc struct {
	Addresses []common.Address
	Accounts  []*RLPAccount
}

type RLPAccount struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash common.Hash
	Storage  [][2]common.Hash
}

type RLPEnv struct {
	Coinbase    common.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes [][2]common.Hash
}

type RLPMessage struct {
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

type RLPResult struct {
	Status uint64
	Bloom  types.Bloom
	Logs   []*types.Log

	ContractAddress common.Address
	GasUsed         uint64
}

// ToRLP decodes val into RLP and returns it.
func ToRLP(val []byte, block uint64) (*RLP, error) {
	var (
		substateRLP RLP
		done        bool
		err         error
	)

	if IsLondonFork(block) {
		err = rlp.DecodeBytes(val, &substateRLP)
		if err == nil {
			return &substateRLP, nil
		}
	}

	if IsBerlinFork(block) && !done {
		var berlin berlinRLP
		err = rlp.DecodeBytes(val, &berlin)
		if err == nil {
			return berlin.toLondon(), nil
		}
	}

	if !done {
		var legacy legacyRLP
		err = rlp.DecodeBytes(val, &legacy)
		if err == nil {
			return legacy.toLondon(), nil
		}

	}

	return nil, err
}

// ----------------------------------------------------------------------------
//                            Legacy Stuff
// ----------------------------------------------------------------------------

// IsLondonFork returns true if block is part of the london fork block range
func IsLondonFork(block uint64) bool {
	return block >= londonBlock
}

// IsBerlinFork returns true if block is part of the berlin fork block range
func IsBerlinFork(block uint64) bool {
	return block >= berlinBlock && block < londonBlock
}

func (r RLP) ToSubstate() (*Substate, error) {
	panic("not yet implemented")
}

// legacyRLP represents legacy RLP structure between before Berlin fork thus before berlinBlock
type legacyRLP struct {
	InputAlloc  RLPAlloc
	OutputAlloc RLPAlloc
	Env         *legacyEnv
	Message     *legacyMessage
	Result      *RLPResult
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

func (m legacyMessage) toLondon() *RLPMessage {
	return &RLPMessage{
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

func (e legacyEnv) toLondon() *RLPEnv {
	return &RLPEnv{
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
	InputAlloc  RLPAlloc
	OutputAlloc RLPAlloc
	Env         *legacyEnv
	Message     *berlinMessage
	Result      *RLPResult
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

func (m berlinMessage) toLondon() *RLPMessage {
	return &RLPMessage{
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
