package types

import (
	"math/big"

	"github.com/holiman/uint256"
)

// BytesToUint256 strictly returns nil if b is nil
func BytesToUint256(b []byte) *uint256.Int {
	if b == nil {
		return nil
	}
	return new(uint256.Int).SetBytes(b)
}

// BytesToBigInt strictly returns nil if b is nil
func BytesToBigInt(b []byte) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}

// BigIntToUint256 strictly returns nil if big is nil
func BigIntToUint256(i *big.Int) *uint256.Int {
	if i == nil {
		return nil
	}
	return uint256.MustFromBig(i)
}
