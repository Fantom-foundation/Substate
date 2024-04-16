package types

import (
	"encoding/json"
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

func TestHash_MarshalText(t *testing.T) {
	m := map[Hash]Hash{
		{1}: {2},
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "{\"0x0100000000000000000000000000000000000000000000000000000000000000\":\"0x0200000000000000000000000000000000000000000000000000000000000000\"}" {
		t.Fatal("incorrect marshalling")
	}
}

func TestAddress_MarshalText(t *testing.T) {
	addr := HexToAddress("0x9c1a711a5e31a9461f6d1f662068e0a2f9edf552")
	m := map[Address]Address{
		addr: addr,
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "{\"0x9c1a711a5e31a9461f6d1f662068e0a2f9edf552\":\"0x9c1a711a5e31a9461f6d1f662068e0a2f9edf552\"}" {
		t.Fatal("incorrect marshalling")
	}
}
