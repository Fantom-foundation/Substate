package rlp

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/rlp"
)

var (
	addr1 = types.Address{0x01}
	hash1 = types.Hash{0x01}
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

func Test_ToSubstateLooksAccountsCodeHashInDatabase(t *testing.T) {
	baseHash := types.Hash{1}
	r := RLP{
		InputSubstate: WorldState{},
		OutputSubstate: WorldState{
			Addresses: []types.Address{addr1},
			Accounts: []*SubstateAccountRLP{{
				Balance:  big.NewInt(10),
				CodeHash: baseHash,
			}},
		},
		Env: &Env{
			BaseFee: new(types.Hash),
		},
		Message: &Message{
			To: new(types.Address),
		},
		Result: &Result{},
	}

	wantedCode := []byte{2}

	ss, err := r.ToSubstate(func(_ types.Hash) ([]byte, error) {
		return wantedCode, nil
	}, 1, 1)
	if err != nil {
		t.Fatalf("cannot convert rlp to substate; %v", err)
	}
	if !bytes.Equal(ss.OutputSubstate[addr1].Code, wantedCode) {
		t.Fatalf("unexpected code was generated\ngot: %s\nwant: %s", string(ss.OutputSubstate[addr1].Code), string(wantedCode))
	}

}

func Test_Message_ToSubstate_CorrectlyAssignsDataIfContractCreation(t *testing.T) {
	r := Message{
		InitCodeHash: &hash1,
	}
	wantedData := []byte{2}
	getHash := func(codeHash types.Hash) ([]byte, error) {
		return wantedData, nil
	}

	m, err := r.ToSubstate(getHash)
	if err != nil {
		t.Fatalf("cannot convert rlp to substate; %v", err)
	}

	if !bytes.Equal(wantedData, m.Data) {
		t.Fatalf("unexpected data\ngot: %v\n want: %v", wantedData, m.Data)
	}
}
