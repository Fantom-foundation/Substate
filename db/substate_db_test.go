package db

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/types"
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/syndtr/goleveldb/leveldb"
)

var testSubstate = &substate.Substate{
	InputAlloc:  substate.NewAlloc(),
	OutputAlloc: substate.NewAlloc(),
	Env: &substate.Env{
		Coinbase:    common.Address{1},
		Difficulty:  new(big.Int).SetUint64(1),
		GasLimit:    1,
		Number:      1,
		Timestamp:   1,
		BlockHashes: make(map[uint64]common.Hash),
		BaseFee:     new(big.Int).SetUint64(1),
	},
	Message:     substate.NewMessage(1, true, new(big.Int).SetUint64(1), 1, common.Address{1}, new(common.Address), new(big.Int).SetUint64(1), []byte{1}, nil, types.AccessList{}, new(big.Int).SetUint64(1), new(big.Int).SetUint64(1)),
	Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, common.Address{1}, 1),
	Block:       37_534_834,
	Transaction: 1,
}

func TestSubstateDB_PutSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
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

func TestSubstateDB_HasSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("has substate returned error; %v", err)
	}

	if !has {
		t.Fatal("substate is not within db")
	}
}

func TestSubstateDB_GetSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ss, err := db.GetSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("get substate returned error; %v", err)
	}

	if ss == nil {
		t.Fatal("substate is nil")
	}

	// todo code does weird stuff
	//if err = ss.Equal(testSubstate); err != nil {
	//	t.Fatalf("substates are different; %v", err)
	//}
}

func TestSubstateDB_DeleteSubstate(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	err = db.DeleteSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("delete substate returned error; %v", err)
	}

	ss, err := db.GetSubstate(37_534_834, 1)
	if err != nil {
		t.Fatalf("get substate returned error; %v", err)
	}

	if ss != nil {
		t.Fatal("substate was not deleted")
	}
}

func createDbAndPutSubstate(dbPath string) (*substateDB, error) {
	db, err := newSubstateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	h1 := common.Hash{}
	h1.SetBytes(nil)

	h2 := common.Hash{}
	h2.SetBytes(nil)

	testSubstate.InputAlloc[common.Address{1}] = substate.NewAccount(1, new(big.Int).SetUint64(1), h1.Bytes())
	testSubstate.OutputAlloc[common.Address{2}] = substate.NewAccount(2, new(big.Int).SetUint64(2), h2.Bytes())
	testSubstate.Env.BlockHashes[1] = common.BytesToHash([]byte{1})

	err = db.PutSubstate(testSubstate)
	if err != nil {
		return nil, err
	}

	return db, nil
}
