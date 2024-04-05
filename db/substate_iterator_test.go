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

func TestSubstateIterator_FromBlock(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(path)
	if err != nil {
		t.Fatal(err)
	}

	test2 := *testSubstate
	test2.Block++

	err = db.PutSubstate(&test2)
	if err != nil {
		t.Fatalf("unable to put substate: %v", err)
	}

	iter := db.NewSubstateIterator(37_534_834, 10)

	if !iter.Next() {
		t.Fatal("next must return true")
	}

	ss := iter.Value()
	if ss.Block != 37_534_834 {
		t.Fatal("incorrect block number")
	}

	counter := 1
	for iter.Next() {
		counter++
	}

	if counter != 2 {
		t.Fatal("incorrect number of substates")
	}

	iter2 := db.NewSubstateIterator(37_534_835, 10)

	if !iter2.Next() {
		t.Fatal("next must return true")
	}

	ss2 := iter2.Value()

	if ss2 == nil {
		t.Fatal("iterator returned nil")
	}

	if ss2.Block != 37_534_835 {
		t.Fatalf("iterator returned transaction with different block number\ngot: %v\n want: %v", ss.Block, 37_534_835)
	}

	if ss2.Transaction != 1 {
		t.Fatalf("iterator returned transaction with different transaction number\ngot: %v\n want: %v", ss2.Transaction, 1)
	}
}
