package rlp

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	substate "github.com/Fantom-foundation/Substate"
	"github.com/Fantom-foundation/Substate/geth/rlp"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	testAddr1 = common.Address{1}
	testAddr2 = common.Address{2}
	testHash1 = common.Hash{1}
	testHash2 = common.Hash{2}
	testHash3 = common.Hash{3}
	testBloom = gethTypes.Bloom{1}
	testLog   = &gethTypes.Log{
		Address:     testAddr1,
		Topics:      []common.Hash{testHash1, testHash2},
		Data:        []byte{1},
		BlockNumber: 10,
		TxHash:      testHash3,
		TxIndex:     20,
		BlockHash:   testHash1,
		Index:       30,
		Removed:     false,
	}
)

func Test_DecodeFromOldLegacy(t *testing.T) {
	legacy := substate.LegacySubstateRLP{
		InputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},
		OutputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},
		Env: &substate.LegacySubstateEnvRLP{
			Coinbase:   testAddr1,
			Difficulty: big.NewInt(1),
			GasLimit:   10,
			Number:     20,
			Timestamp:  30,
			BlockHashes: [][2]common.Hash{
				{testHash1, testHash2}, {testHash3},
			},
		},
		Message: &substate.LegacySubstateMessageRLP{
			Nonce:      10,
			CheckNonce: false,
			GasPrice:   big.NewInt(1),
			Gas:        10,
			From:       testAddr1,
			To:         &testAddr2,
			Value:      big.NewInt(1),
			Data:       []byte{1},
		},
		Result: &substate.SubstateResultRLP{
			Status:          1,
			Bloom:           testBloom,
			Logs:            []*gethTypes.Log{testLog},
			ContractAddress: testAddr2,
			GasUsed:         10,
		},
	}
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

func Test_DecodeFromOldBerlin(t *testing.T) {
	berlin := substate.BerlinSubstateRLP{
		InputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},
		OutputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},

		Env: &substate.LegacySubstateEnvRLP{
			Coinbase:   testAddr1,
			Difficulty: big.NewInt(1),
			GasLimit:   10,
			Number:     20,
			Timestamp:  30,
			BlockHashes: [][2]common.Hash{
				{testHash1, testHash2}, {testHash3},
			},
		},
		Message: &substate.BerlinSubstateMessageRLP{
			Nonce:      10,
			CheckNonce: false,
			GasPrice:   big.NewInt(1),
			Gas:        10,
			From:       testAddr1,
			To:         &testAddr2,
			Value:      big.NewInt(1),
			Data:       []byte{1},
			AccessList: gethTypes.AccessList{gethTypes.AccessTuple{Address: testAddr1, StorageKeys: []common.Hash{testHash1, testHash2}}},
		},
		Result: &substate.SubstateResultRLP{
			Status:          1,
			Bloom:           testBloom,
			Logs:            []*gethTypes.Log{testLog},
			ContractAddress: testAddr2,
			GasUsed:         10,
		},
	}
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

func Test_DecodeFromOldLondon(t *testing.T) {
	london := substate.SubstateRLP{
		InputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},
		OutputAlloc: substate.SubstateAllocRLP{
			Addresses: []common.Address{testAddr1, testAddr2},
			Accounts: []*substate.SubstateAccountRLP{
				{
					Nonce:    10,
					Balance:  big.NewInt(1),
					CodeHash: testHash1,
					Storage: [][2]common.Hash{
						{testHash1, testHash2}, {testHash3},
					},
				},
				{
					Nonce:    20,
					Balance:  big.NewInt(2),
					CodeHash: testHash2,
					Storage: [][2]common.Hash{
						{testHash2, testHash3}, {testHash1},
					},
				},
			},
		},

		Message: &substate.SubstateMessageRLP{
			Nonce:      10,
			CheckNonce: false,
			GasPrice:   big.NewInt(1),
			Gas:        10,
			From:       testAddr1,
			To:         &testAddr2,
			Value:      big.NewInt(1),
			Data:       []byte{1},
			AccessList: gethTypes.AccessList{gethTypes.AccessTuple{Address: testAddr1, StorageKeys: []common.Hash{testHash1, testHash2}}},
			GasFeeCap:  big.NewInt(1),
			GasTipCap:  big.NewInt(1),
		},
		Env: &substate.SubstateEnvRLP{
			Coinbase:   testAddr1,
			Difficulty: big.NewInt(1),
			GasLimit:   10,
			Number:     20,
			Timestamp:  30,
			BlockHashes: [][2]common.Hash{
				{testHash1, testHash2}, {testHash3},
			},
		},
		Result: &substate.SubstateResultRLP{
			Status:          1,
			Bloom:           testBloom,
			Logs:            []*gethTypes.Log{testLog},
			ContractAddress: testAddr2,
			GasUsed:         10,
		},
	}
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
