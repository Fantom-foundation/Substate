package new_substate

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/types"
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

func (e *Env) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Coinbase: %v\n", e.Coinbase.Hex()))
	builder.WriteString(fmt.Sprintf("Difficulty: %v\n", e.Difficulty.String()))
	builder.WriteString(fmt.Sprintf("Gas Limit: %v\n", e.GasLimit))
	builder.WriteString(fmt.Sprintf("Number: %v\n", e.Number))
	builder.WriteString(fmt.Sprintf("Timestamp: %v\n", e.Timestamp))
	builder.WriteString(fmt.Sprintf("Base Fee: %v\n", e.BaseFee.String()))
	builder.WriteString("Block Hashes: \n")

	for number, hash := range e.BlockHashes {
		builder.WriteString(fmt.Sprintf("%v: %v\n", number, hash.Hex()))
	}

	return builder.String()

}
