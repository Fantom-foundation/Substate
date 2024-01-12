package substate

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/geth/common"
)

func TestAlloc_Add(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}
	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	alloc := make(Alloc).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	alloc.Add(addr2, acc.Nonce, acc.Balance, acc.Code)

	if len(alloc) != 2 {
		t.Fatalf("incorrect len after add\n got: %v\n want: %v", len(alloc), 2)
	}

	if !alloc[addr2].Equal(acc) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", alloc[addr2], acc)
	}
}

func TestAlloc_MergeOneAccount(t *testing.T) {
	addr := common.Address{1}

	alloc := make(Alloc).Add(addr, 1, new(big.Int).SetUint64(1), []byte{1})
	allocToMerge := make(Alloc).Add(addr, 2, new(big.Int).SetUint64(2), []byte{2})

	alloc.Merge(allocToMerge)

	acc := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !alloc[addr].Equal(acc) {
		t.Fatalf("incorrect merge\ngot: %v\n want: %v", alloc[addr], acc)
	}

}

func TestAlloc_MergeTwoAccounts(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}

	alloc := make(Alloc).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	allocToMerge := make(Alloc).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	alloc.Merge(allocToMerge)

	want1 := &Account{
		Nonce:   1,
		Balance: new(big.Int).SetUint64(1),
		Code:    []byte{1},
	}

	if len(alloc) != 2 {
		t.Fatalf("incorrect len after merge\n got: %v\n want: %v", len(alloc), 2)
	}

	if !alloc[addr1].Equal(want1) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", alloc[addr1], want1)
	}

	want2 := &Account{
		Nonce:   2,
		Balance: new(big.Int).SetUint64(2),
		Code:    []byte{2},
	}

	if !alloc[addr2].Equal(want2) {
		t.Fatalf("incorrect merge of addr1\ngot: %v\n want: %v", alloc[addr2], want2)
	}

}

func TestAlloc_EstimateIncrementalSize_NewAlloc(t *testing.T) {
	addr1 := common.Address{1}
	addr2 := common.Address{2}

	alloc := make(Alloc).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	allocToEstimate := make(Alloc).Add(addr2, 2, new(big.Int).SetUint64(2), []byte{2})

	want := sizeOfAddress + sizeOfNonce + uint64(len(allocToEstimate[addr2].Balance.Bytes())) + sizeOfHash

	// adding new alloc without storage keys
	if got := alloc.EstimateIncrementalSize(allocToEstimate); got != want {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, want)
	}
}

func TestAlloc_EstimateIncrementalSize_SameAlloc(t *testing.T) {
	addr1 := common.Address{1}

	alloc := make(Alloc).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	allocToEstimate := make(Alloc).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})

	// since we don't add anything, size should not be increased
	if got := alloc.EstimateIncrementalSize(allocToEstimate); got != 0 {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, 0)
	}
}

func TestAlloc_EstimateIncrementalSize_AddingStorageHash(t *testing.T) {
	addr1 := common.Address{1}

	alloc := make(Alloc).Add(addr1, 1, new(big.Int).SetUint64(1), []byte{1})
	allocToEstimate := make(Alloc).Add(addr1, 2, new(big.Int).SetUint64(2), []byte{2})
	allocToEstimate[addr1].Storage[common.Hash{1}] = common.Hash{1}

	// we add one key to already existing account, this size is increased by the sizeOfHash
	if got := alloc.EstimateIncrementalSize(allocToEstimate); got != sizeOfHash {
		t.Fatalf("incorrect estimation\ngot: %v\nwant: %v", got, sizeOfHash)
	}
}

// todo diff tests

func TestAlloc_Equal(t *testing.T) {
	alloc := make(Alloc).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedAllocEqual := make(Alloc).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})

	if !alloc.Equal(comparedAllocEqual) {
		t.Fatal("allocs are same but equal returned false")
	}
}

func TestAlloc_NotEqual(t *testing.T) {
	alloc := make(Alloc).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedAllocEqual := make(Alloc).Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if alloc.Equal(comparedAllocEqual) {
		t.Fatal("allocs are different but equal returned false")
	}
}

func TestAlloc_NotEqual_DifferentLen(t *testing.T) {
	alloc := make(Alloc).Add(common.Address{1}, 1, new(big.Int).SetUint64(1), []byte{1})
	comparedAllocEqual := make(Alloc).Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	// add one more acc to alloc
	alloc.Add(common.Address{2}, 1, new(big.Int).SetUint64(1), []byte{1})

	if alloc.Equal(comparedAllocEqual) {
		t.Fatal("allocs are different but equal returned false")
	}
}

func TestAlloc_Copy(t *testing.T) {
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
