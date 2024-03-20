package common

import (
	"encoding/hex"
	"math/big"
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [32]byte

func (h Hash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

func (h Hash) IsEmpty() bool {
	return h == Hash{}
}

// Uint64 converts a hash to a uint64.
func (h Hash) Uint64() uint64 { return new(big.Int).SetBytes(h[:]).Uint64() }

// Compare two big int representations of h and h2.
func (h Hash) Compare(h2 Hash) int {
	b1 := new(big.Int).SetBytes(h[:])
	b2 := new(big.Int).SetBytes(h2[:])
	return b1.Cmp(b2)
}

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-32:]
	}

	copy(h[32-len(b):], b)
}

// BigToHash sets byte representation of b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }
