package types

import (
	"github.com/holiman/uint256"
)

// BytesToUint256 in research package strictly returns nil if b is nil
func BytesToUint256(b []byte) *uint256.Int {
	if b == nil {
		return nil
	}
	return uint256.MustFromBig(BytesToBigInt(b))
}

// BytesToBigInt in research package strictly returns nil if b is nil
func BytesToBigInt(b []byte) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}
