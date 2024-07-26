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
		BlockHashes: createBlockHashes(env.BlockHashes),
	}

	e.BaseFee = nil
	if env.BaseFee != nil {
		baseFee := types.BigToHash(env.BaseFee)
		e.BaseFee = &baseFee
	}

	e.BlobBaseFee = nil
	if env.BlobBaseFee != nil {
		blobBaseFee := types.BigToHash(env.BlobBaseFee)
		e.BlobBaseFee = &blobBaseFee
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

	BaseFee     *types.Hash `rlp:"nil"` // missing in substate DB from Geth <= v1.10.3
	BlobBaseFee *types.Hash `rlp:"nil"` // missing in substate DB before Cancun
}

// ToSubstate transforms e from Env to substate.Env.
func (e Env) ToSubstate() *substate.Env {
	var baseFee, blobBaseFee *big.Int
	if e.BaseFee != nil {
		baseFee = e.BaseFee.Big()
	}

	if e.BlobBaseFee != nil {
		blobBaseFee = e.BlobBaseFee.Big()
	}

	se := &substate.Env{
		Coinbase:    e.Coinbase,
		Difficulty:  e.Difficulty,
		GasLimit:    e.GasLimit,
		Number:      e.Number,
		Timestamp:   e.Timestamp,
		BlockHashes: make(map[uint64]types.Hash),
		BaseFee:     baseFee,
		BlobBaseFee: blobBaseFee,
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
