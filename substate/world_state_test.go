package substate

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/types/common"
)

func TestWorldState_Add(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}
	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldState.Add(addr2, acc.Nonce, acc.Balance, acc.Code)

	if len(worldState) != 2 {
		t.Fatalf("incorrect len after add\n got: %v\n want: %v", len(worldState), 2)
	}

	if !worldState[addr2].Equal(acc) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr2], acc)
	}
}

func TestWorldState_MergeOneAccount(t *testing.T) {
	addr := common.Address{1}

	worldState := make(WorldState).Add(addr, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToMerge := make(WorldState).Add(addr, 2, new(big.Int).SetUint64(2), []byte{2})

	worldState.Merge(worldStateToMerge)

	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !worldState[addr].Equal(acc) {
		t.Fatalf("incorrect merge\ngot: %v\n want: %v", worldState[addr], acc)
	}

}

func TestWorldState_MergeTwoAccounts(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToMerge := make(WorldState).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	worldState.Merge(worldStateToMerge)

	want1 := &Account{
		Nonce:   1,
		Balance: new(big.Int).SetUint64(1),
		Code:    []byte{1},
	}

	if len(worldState) != 2 {
		t.Fatalf("incorrect len after merge\n got: %v\n want: %v", len(worldState), 2)
	}

	if !worldState[addr1].Equal(want1) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr1], want1)
	}

	want2 := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !worldState[addr2].Equal(want2) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", worldState[addr2], want2)
	}

}

func TestWorldState_EstimateIncrementalSize_NewWorldState(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	want := sizeOfAddress + sizeOfNonce + uint64(len(worldStateToEstimate[addr2].Balance.Bytes())) + sizeOfHash

	// adding new world state without storage keys
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != want {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, want)
	}
}

func TestWorldState_EstimateIncrementalSize_SameWorldState(t *testing.T) {
	addr1 := common.Address{1}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})

	// since we don't add anything, size should not be increased
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != 0 {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, 0)
	}
}

func TestWorldState_EstimateIncrementalSize_AddingStorageHash(t *testing.T) {
	addr1 := common.Address{1}

	worldState := make(WorldState).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	worldStateToEstimate := make(WorldState).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})
	worldStateToEstimate[addr1].Storage[common.Hash{1}] = common.Hash{1}

	// we add one key to already existing account, this size is increased by the sizeOfHash
	if got := worldState.EstimateIncrementalSize(worldStateToEstimate); got != sizeOfHash {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, sizeOfHash)
	}
}

// todo diff tests

func TestWorldState_Equal(t *testing.T) {
	worldState := make(WorldState).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})

	if !worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are same but equal returned false")
	}
}

func TestWorldState_NotEqual(t *testing.T) {
	worldState := make(WorldState).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are different but equal returned false")
	}
}

func TestWorldState_NotEqual_DifferentLen(t *testing.T) {
	worldState := make(WorldState).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedWorldStateEqual := make(WorldState).Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	// add one more acc to world state
	worldState.Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if worldState.Equal(comparedWorldStateEqual) {
		t.Fatal("world states are different but equal returned false")
	}
}

func TestWorl_Copy(t *testing.T) {
	hashOne := common.BigToHash(new(big.Int).SetUint64(1))
	hashTwo := common.BigToHash(new(big.Int).SetUint64(2))
	acc := NewAccount(1, new(big.Int).SetUint64(1), []byte{1})
	acc.Storage = make(map[common.Hash]common.Hash)
	acc.Storage[hashOne] = hashTwo

	cpy := acc.Copy()
	if !acc.Equal(cpy) {
		t.Fatalf("accounts values must be equal\ngot: %v\nwant: %v", cpy, acc)
	}
}
