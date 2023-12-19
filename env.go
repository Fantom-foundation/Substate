package substate

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Env struct {
	Coinbase    common.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes map[uint64]common.Hash

	// London hard fork, EIP-1559
	BaseFee *big.Int // nil if EIP-1559 is not activated
}

func NewEnv(b *types.Block, blockHashes map[uint64]common.Hash) *Env {
	return &Env{
		Coinbase:    b.Coinbase(),
		Difficulty:  new(big.Int).Set(b.Difficulty()),
		GasLimit:    b.GasLimit(),
		Number:      b.NumberU64(),
		Timestamp:   b.Time(),
		BlockHashes: blockHashes,
		BaseFee:     b.BaseFee(),
	}
}
