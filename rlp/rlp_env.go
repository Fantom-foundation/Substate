package rlp

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
)

func NewEnv(env *substate.Env) *Env {
	e := &Env{
		Coinbase:    env.Coinbase,
		Difficulty:  env.Difficulty,
		GasLimit:    env.GasLimit,
		Number:      env.Number,
		Timestamp:   env.Timestamp,
		BlockHashes: nil,
	}

	var sortedNum64 []uint64
	for num64 := range env.BlockHashes {
		sortedNum64 = append(sortedNum64, num64)
	}

	for _, num64 := range sortedNum64 {
		num := types.BigToHash(new(big.Int).SetUint64(num64))
		blockHash := env.BlockHashes[num64]
		pair := [2]types.Hash{num, blockHash}
		e.BlockHashes = append(e.BlockHashes, pair)
	}

	e.BaseFee = nil
	if env.BaseFee != nil {
		baseFeeHash := types.BigToHash(env.BaseFee)
		e.BaseFee = &baseFeeHash
	}

	return e
}

type Env struct {
	Coinbase    types.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes [][2]types.Hash

	BaseFee *types.Hash `rlp:"nil"` // missing in substate DB from Geth <= v1.10.3
}

// ToSubstate transforms e from Env to substate.Env.
func (e Env) ToSubstate() *substate.Env {
	var baseFee *big.Int
	if e.BaseFee != nil {
		baseFee = e.BaseFee.Big()
	}

	se := &substate.Env{
		Coinbase:    e.Coinbase,
		Difficulty:  e.Difficulty,
		GasLimit:    e.GasLimit,
		Number:      e.Number,
		Timestamp:   e.Timestamp,
		BlockHashes: make(map[uint64]types.Hash),
		BaseFee:     baseFee,
	}

	// iterate through BlockHashes
	// first hash is the block number
	// second hash is the block hash itself
	for _, hashes := range e.BlockHashes {
		number := hashes[0].Uint64()
		hash := hashes[1]
		se.BlockHashes[number] = hash
	}

	return se

}
