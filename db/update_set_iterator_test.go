package db

import (
	"testing"
	"time"
)

func TestUpdateSetIterator_Next(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)
	if !iter.Next() {
		t.Fatal("next must return true")
	}

	if iter.Next() {
		t.Fatal("next must return false, all update-sets were extracted")
	}
}

func TestUpdateSetIterator_Value(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	tx := iter.Value()

	if tx == nil {
		t.Fatal("iterator returned nil")
	}

	if tx.Block != 1 {
		t.Fatalf("iterator returned UpdateSet with different block number\ngot: %v\n want: %v", tx.Block, 1)
	}

}

func TestUpdateSetIterator_Release(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutUpdateSet(path)
	if err != nil {
		return
	}

	iter := db.NewUpdateSetIterator(0, 10)

	// make sure Release is not blocking.
	done := make(chan bool)
	go func() {
		iter.Release()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Second):
		t.Fatal("Release blocked unexpectedly")
	}

}
