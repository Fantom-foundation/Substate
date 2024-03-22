package rlp

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/geth/rlp"
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

	res, err := Decode(b, londonBlock)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}

func Test_DecodeLondon_IncorrectDecoder(t *testing.T) {
	london := RLP{
		Message: &Message{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &Env{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(london)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decode(b, berlinBlock)
	if err == nil {
		t.Fatal(err)
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

	res, err := Decode(b, berlinBlock)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}

func Test_DecodeBerlin_IncorrectDecoder(t *testing.T) {
	berlin := berlinRLP{
		Message: &berlinMessage{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &legacyEnv{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(berlin)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decode(b, londonBlock)
	if err == nil {
		t.Fatal(err)
	}
}

func Test_DecodeLegacy(t *testing.T) {
	legacy := legacyRLP{
		Message: &legacyMessage{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &legacyEnv{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(legacy)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Decode(b, berlinBlock-1)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res.Message.Data, []byte{1}) {
		t.Fatal("incorrect data")
	}
}

func Test_DecodeLegacy_IncorrectDecoder(t *testing.T) {
	legacy := legacyRLP{
		Message: &legacyMessage{Data: []byte{1}, Value: big.NewInt(1), GasPrice: big.NewInt(1)},
		Env:     &legacyEnv{},
		Result:  &Result{}}
	b, err := rlp.EncodeToBytes(legacy)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decode(b, londonBlock)
	if err == nil {
		t.Fatal(err)
	}
}
