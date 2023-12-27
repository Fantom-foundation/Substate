package db

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Fantom-foundation/Substate/geth/crypto"
	"github.com/syndtr/goleveldb/leveldb"
)

var testCode = []byte{1}

func TestCodeDB_PutCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"

	db, err := createDbAndPutCode(dbPath)
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

func TestCodeDB_HasCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	has, err := db.HasCode(crypto.Keccak256Hash(testCode))
	if err != nil {
		t.Fatalf("get code returned error; %v", err)
	}

	if !has {
		t.Fatal("code is not within db")
	}
}

func TestCodeDB_GetCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	code, err := db.GetCode(crypto.Keccak256Hash(testCode))
	if err != nil {
		t.Fatalf("get code returned error; %v", err)
	}

	if bytes.Compare(code, testCode) != 0 {
		t.Fatal("code returned by the db is different")
	}
}

func TestCodeDB_DeleteCode(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutCode(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256Hash(testCode)

	err = db.DeleteCode(hash)
	if err != nil {
		t.Fatalf("delete code returned error; %v", err)
	}

	code, err := db.GetCode(hash)
	if err != nil {
		t.Fatalf("get code returned error; %v", err)
	}

	if code != nil {
		t.Fatal("code was not deleted")
	}
}

func createDbAndPutCode(dbPath string) (*codeDB, error) {
	db, err := newCodeDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open db; %v", err)
	}

	err = db.PutCode(testCode)
	if err != nil {
		return nil, err
	}

	return db, nil
}
