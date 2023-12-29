package db

import (
	"testing"
	"time"
)

func TestSubstateIterator_Next(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	iter := db.NewSubstateIterator(0, 10)
	if !iter.Next() {
		t.Fatal("next must return true")
	}

	if iter.Next() {
		t.Fatal("next must return false, all substates were extracted")
	}
}

func TestSubstateIterator_Value(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	iter := db.NewSubstateIterator(0, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	tx := iter.Value()

	if tx == nil {
		t.Fatal("iterator returned nil")
	}

	if tx.Block != 37_534_834 {
		t.Fatalf("iterator returned transaction with different block number\ngot: %v\n want: %v", tx.Block, 37_534_834)
	}

	if tx.Transaction != 1 {
		t.Fatalf("iterator returned transaction with different transaction number\ngot: %v\n want: %v", tx.Transaction, 1)
	}
}

func TestSubstateIterator_Release(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		return
	}

	iter := db.NewSubstateIterator(0, 10)

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
