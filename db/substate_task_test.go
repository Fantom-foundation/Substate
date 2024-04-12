package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/Substate/substate"
)

func TestSubstateTaskPool_Execute(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// add one more substate
	if err = addSubstate(db, testSubstate.Block+1); err != nil {
		t.Fatal(err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		Workers: 1,
		DB:      db,
	}

	err = stPool.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubstateTaskPool_ExecuteBlock(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(testSubstate.Block)
	require.Nil(t, err)
	require.Equal(t, int64(1), numTx)
	require.Equal(t, testSubstate.Message.Gas, uint64(gas))
}

func TestSubstateTaskPool_ExecuteBlock_TaskFuncErr(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("test error")
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		Workers: 1,
		DB:      db,
	}

	_, _, err = stPool.ExecuteBlock(testSubstate.Block)
	require.Error(t, err)
}

func TestSubstateTaskPool_ExecuteBlockNilTaskFunc(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(testSubstate.Block)
	require.Nil(t, err)
	require.Equal(t, int64(1), numTx)
	require.Equal(t, int64(0), gas)
}

func TestSubstateTaskPool_ExecuteBlockDBError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := newSubstateDB(dbPath, nil, nil, nil)
	if err != nil {
		t.Fatalf("cannot open db; %v", err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		BlockFunc: func(block uint64, transactions map[int]*substate.Substate, taskPool *SubstateTaskPool) error {
			return errors.New("test error")
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		Workers: 1,
		DB:      db,
	}

	_, _, err = stPool.ExecuteBlock(testSubstate.Block)
	require.Error(t, err)
}

func TestSubstateTaskPool_ExecuteBlockSkipTransferTx(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := createDbAndPutSubstate(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	stPool := SubstateTaskPool{
		Name: "test",

		TaskFunc: func(block uint64, tx int, substate *substate.Substate, taskPool *SubstateTaskPool) error {
			return nil
		},

		First: testSubstate.Block,
		Last:  testSubstate.Block + 1,

		SkipTransferTxs: true,

		Workers: 1,
		DB:      db,
	}

	numTx, gas, err := stPool.ExecuteBlock(testSubstate.Block)
	require.Nil(t, err)
	require.Equal(t, int64(0), numTx)
	require.Equal(t, int64(0), gas)
}
