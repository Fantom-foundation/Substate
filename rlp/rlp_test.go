package rlp

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/types/rlp"
)

func Test_DecodeLondon(t *testing.T) {
	london := RLP{
		Message: &Message{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &Env{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(london)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}

func Test_DecodeBerlin(t *testing.T) {
	berlin := berlinRLP{
		Message: &berlinMessage{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &legacyEnv{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(berlin)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}

func Test_DecodeLegacy(t *testing.T) {
	legacy := legacySubstateRLP{
		Message: &legacyMessage{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &legacyEnv{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(legacy)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}
