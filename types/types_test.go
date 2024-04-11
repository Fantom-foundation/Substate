package types

import (
	"math/big"
	"testing"
)

func TestAddress_Convertation(t *testing.T) {
	addrStr := "0x9c1a711a5e31a9461f6d1f662068e0a2f9edf552"
	addr := HexToAddress(addrStr)

	if addr.String() != addrStr {
		t.Fatal("incorrect address conversion")
	}
}

func TestHash_Compare(t *testing.T) {
	a := big.NewInt(1)
	b := big.NewInt(2)

	h1 := BigToHash(a)
	h2 := BigToHash(b)

	if h1.Compare(h2) != -1 {
		t.Fatal("incorrect comparing")
	}

	if h1.Uint64() != uint64(1) {
		t.Fatal("incorrect uint64 conversion")
	}
}
