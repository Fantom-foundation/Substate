package db

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/updateset"
)

var testUpdateSet = &updateset.UpdateSet{
	WorldState: substate.WorldState{
		types.Address{1}: &substate.Account{
			Nonce:   1,
			Balance: new(big.Int).SetUint64(1),
		},
		types.Address{2}: &substate.Account{
			Nonce:   2,
			Balance: new(big.Int).SetUint64(2),
		},
	},
	Block: 1,
}

var testDeletedAccounts = []types.Address{{3}, {4}}

func TestUpdateDB_PutUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	s := new(leveldb.DBStats)
	err = db.backend.Stats(s)
	if err != nil {
		t.Fatalf("cannot get db stats; %v", err)
	}

	// 54 is the base write when creating levelDB
	if s.IOWrite <= 54 {
		t.Fatal("db file should have something inside")
	}
}

func TestUpdateDB_HasUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasUpdateSet(testUpdateSet.Block)
	if err != nil {
		t.Fatalf("has update-set returned error; %v", err)
	}

	if !has {
		t.Fatal("update-set is not within db")
	}
}

func TestUpdateDB_GetUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	us, err := db.GetUpdateSet(testUpdateSet.Block)
	if err != nil {
		t.Fatalf("get update-set returned error; %v", err)
	}

	if us == nil {
		t.Fatal("update-set is nil")
	}

	if !us.Equal(testUpdateSet) {
		t.Fatal("substates are different")
	}
}

func TestUpdateDB_DeleteUpdateSet(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	err = db.DeleteUpdateSet(testUpdateSet.Block)
	if err != nil {
		t.Fatalf("delete update-set returned error; %v", err)
	}

	us, err := db.GetUpdateSet(testUpdateSet.Block)
	if err == nil {
		t.Fatal("get update-set must fail")
	}

	if got, want := err, leveldb.ErrNotFound; !errors.Is(got, want) {
		t.Fatalf("unexpected err, got: %v, want: %v", got, want)
	}

	if us != nil {
		t.Fatal("update-set was not deleted")
	}
}

func TestUpdateDB_GetFirstKey(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetFirstKey()
	if err != nil {
		t.Fatalf("cannot get first key; %v", err)
	}

	var want = testUpdateSet.Block

	if want != got {
		t.Fatalf("incorrect first key\nwant: %v\ngot: %v", want, got)
	}
}

func TestUpdateDB_GetLastKey(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetLastKey()
	if err != nil {
		t.Fatalf("cannot get last key; %v", err)
	}

	var want = testUpdateSet.Block

	if want != got {
		t.Fatalf("incorrect last key\nwant: %v\ngot: %v", want, got)
	}
}

func createDbAndPutUpdateSet(dbPath string) (*updateDB, error) {
	db, err := newUpdateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	err = db.PutUpdateSet(testUpdateSet, testDeletedAccounts)
	if err != nil {
		return nil, err
	}

	return db, nil
}
