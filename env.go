package substate

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/common"
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
		//Coinbase:    b.Coinbase(), // todo uncomment when all things are imported from eth
		Difficulty:  new(big.Int).Set(b.Difficulty()),
		GasLimit:    b.GasLimit(),
		Number:      b.NumberU64(),
		Timestamp:   b.Time(),
		BlockHashes: blockHashes,
		BaseFee:     b.BaseFee(),
	}
}

// Equal returns true if e is y or if values of e are equal to values of y.
// Otherwise, e and y are not equal hence false is returned.
func (e *Env) Equal(y *Env) bool {
	if e == y {
		return true
	}

	if (e == nil || y == nil) && e != y {
		return false
	}

	equal := e.Coinbase == y.Coinbase &&
		e.Difficulty.Cmp(y.Difficulty) == 0 &&
		e.GasLimit == y.GasLimit &&
		e.Number == y.Number &&
		e.Timestamp == y.Timestamp &&
		len(e.BlockHashes) == len(y.BlockHashes) &&
		e.BaseFee.Cmp(y.BaseFee) == 0
	if !equal {
		return false
	}

	for k, xv := range e.BlockHashes {
		yv, exist := y.BlockHashes[k]
		if !(exist && xv == yv) {
			return false
		}
	}

	return true
}
