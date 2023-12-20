package substate

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/Substate/common"
	geth "github.com/ethereum/go-ethereum/common"
)

// These tests test whether newly created functions and types correspond with geth based types.

func TestAccount_Account(t *testing.T) {
	nonce := rand.Uint64()
	balance := new(big.Int).SetUint64(rand.Uint64())
	storage, gethStorage := createTestStorages()
	code := []byte{byte(rand.Uint64())}

	acc, gethAcc := createTestAccounts(nonce, balance, storage, gethStorage, code)
	err := compareAccounts(acc, gethAcc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccount_NewAccount(t *testing.T) {
	nonce := rand.Uint64()
	balance := new(big.Int).SetUint64(rand.Uint64())
	code := []byte{byte(rand.Uint64())}

	acc := NewAccount(nonce, balance, code)
	gethAcc := NewSubstateAccount(nonce, balance, code)

	err := compareAccounts(acc, gethAcc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccount_Equal(t *testing.T) {
	nonce := rand.Uint64()
	balance := new(big.Int).SetUint64(rand.Uint64())
	storage, gethStorage := createTestStorages()
	code := []byte{byte(rand.Uint64())}

	// accounts are same
	accX, gethAccX := createTestAccounts(nonce, balance, storage, gethStorage, code)
	accY, gethAccY := createTestAccounts(nonce, balance, storage, gethStorage, code)

	accBool := accX.Equal(accY)
	gethBool := gethAccX.Equal(gethAccY)

	if accBool != gethBool {
		t.Fatalf("equal returned different value\ngot: %v\nwant:%v", accBool, gethBool)
	}

	// accounts have different nonce
	accX, gethAccX = createTestAccounts(nonce+1, balance, storage, gethStorage, code)
	accY, gethAccY = createTestAccounts(nonce+1, balance, storage, gethStorage, code)

	accBool = accX.Equal(accY)
	gethBool = gethAccX.Equal(gethAccY)
	if accBool != gethBool {
		t.Fatalf("equal (nonce) returned different value\ngot: %v\nwant:%v", accBool, gethBool)
	}

	// accounts have different balance
	accX, gethAccX = createTestAccounts(nonce, balance.Add(balance, new(big.Int).SetUint64(1)), storage, gethStorage, code)
	accY, gethAccY = createTestAccounts(nonce, balance, storage, gethStorage, code)
	accBool = accX.Equal(accY)
	gethBool = gethAccX.Equal(gethAccY)
	if accBool != gethBool {
		t.Fatalf("equal (balance) returned different value\ngot: %v\nwant:%v", accBool, gethBool)
	}

	// accounts have different storages
	accX, gethAccX = createTestAccounts(nonce, balance, nil, nil, code)
	accY, gethAccY = createTestAccounts(nonce, balance, storage, gethStorage, code)
	accBool = accX.Equal(accY)
	gethBool = gethAccX.Equal(gethAccY)
	if accBool != gethBool {
		t.Fatalf("equal (storage) returned different value\ngot: %v\nwant:%v", accBool, gethBool)
	}

	// accounts have different codes
	accX, gethAccX = createTestAccounts(nonce, balance, storage, gethStorage, []byte{123})
	accY, gethAccY = createTestAccounts(nonce, balance, storage, gethStorage, code)
	accBool = accX.Equal(accY)
	gethBool = gethAccX.Equal(gethAccY)
	if accBool != gethBool {
		t.Fatalf("equal (code) returned different value\ngot: %v\nwant:%v", accBool, gethBool)
	}

}

// compareAccounts compares values of new Account type with old SubstateAccount type.
func compareAccounts(acc *Account, gethAcc *SubstateAccount) error {
	if acc.Nonce != gethAcc.Nonce {
		return fmt.Errorf("nonce is different\ngot: %v\n want: %v", acc.Nonce, gethAcc.Nonce)
	}

	if acc.Balance.Uint64() != gethAcc.Balance.Uint64() {
		return fmt.Errorf("balance is different\ngot: %v\n want: %v", acc.Balance.Uint64(), gethAcc.Balance.Uint64())
	}

	for key, val := range acc.Storage {
		gethVal, ok := gethAcc.Storage[geth.BytesToHash(key.Bytes())]
		if !ok {
			return fmt.Errorf("hash is not contained in gethStorage\nkey: %v\n value: %v", key.Hex(), val.Hex())
		}

		if bytes.Compare(val.Bytes(), gethVal.Bytes()) != 0 {
			return fmt.Errorf("hash is different\ngot: %v\n want: %v", val.Hex(), gethVal.Hex())
		}
	}

	if bytes.Compare(acc.Code, gethAcc.Code) != 0 {
		return fmt.Errorf("code is different\ngot: %v\n want: %v", string(acc.Code), string(gethAcc.Code))
	}

	return nil
}

func createTestAccounts(nonce uint64, balance *big.Int, storage map[common.Hash]common.Hash, gethStorage map[geth.Hash]geth.Hash, code []byte) (*Account, *SubstateAccount) {

	acc := Account{
		Nonce:   nonce,
		Balance: balance,
		Storage: storage,
		Code:    code,
	}

	gethAcc := SubstateAccount{
		Nonce:   nonce,
		Balance: balance,
		Storage: gethStorage,
		Code:    code,
	}

	return &acc, &gethAcc
}

// createTestStorages one from created common library and one from geth common library
func createTestStorages() (map[common.Hash]common.Hash, map[geth.Hash]geth.Hash) {
	bigOne := new(big.Int).SetUint64(rand.Uint64())
	bigTwo := new(big.Int).SetUint64(rand.Uint64())
	bigThree := new(big.Int).SetUint64(rand.Uint64())
	bigFour := new(big.Int).SetUint64(rand.Uint64())

	hash1 := common.BigToHash(bigOne)
	hash2 := common.BigToHash(bigTwo)
	hash3 := common.BigToHash(bigThree)
	hash4 := common.BigToHash(bigFour)

	storage := make(map[common.Hash]common.Hash)
	storage[hash1] = hash2
	storage[hash2] = hash3
	storage[hash3] = hash4
	storage[hash4] = hash1

	gethHash1 := geth.BigToHash(bigOne)
	gethHash2 := geth.BigToHash(bigTwo)
	gethHash3 := geth.BigToHash(bigThree)
	gethHash4 := geth.BigToHash(bigFour)

	gethStorage := make(map[geth.Hash]geth.Hash)
	gethStorage[gethHash1] = gethHash2
	gethStorage[gethHash2] = gethHash3
	gethStorage[gethHash3] = gethHash4
	gethStorage[gethHash4] = gethHash1

	return storage, gethStorage
}
