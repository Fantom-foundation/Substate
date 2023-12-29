package db

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Fantom-foundation/Substate/new_substate"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// SubstateIterator iterates over a database's key/value pairs in ascending key order.
//
// When it encounters an error any seek will return false and will yield no key/
// value pairs. The error can be queried by calling the Error method. Calling
// Release is still necessary.
//
// An iterator must be released after use, but it is not necessary to read an
// iterator until exhaustion. An iterator is not safe for concurrent use, but it
// is safe to use multiple iterators concurrently.
type SubstateIterator interface {
	// Next moves the iterator to the next key/value pair. It returns whether the
	// iterator is exhausted.
	Next() bool

	// Error returns any accumulated error. Exhausting all the key/value pairs
	// is not considered to be an error.
	Error() error

	// Value returns the value of the current Transaction, or nil if done. The
	// caller should not modify the contents of the returned slice, and its contents
	// may change on the next call to Next.
	Value() *Transaction

	// Release releases associated resources. Release should always succeed and can
	// be called multiple times without causing error.
	Release()
}

func newSubstateIterator(db *substateDB, start []byte) *substateIterator {
	r := util.BytesPrefix([]byte(Stage1SubstatePrefix))
	r.Start = append(r.Start, start...)

	return &substateIterator{
		db:       db,
		iter:     db.backend.NewIterator(r, db.ro),
		resultCh: make(chan *Transaction, 10),
		wg:       new(sync.WaitGroup),
	}
}

type substateIterator struct {
	err      error
	db       *substateDB
	iter     iterator.Iterator
	resultCh chan *Transaction
	wg       *sync.WaitGroup
	cur      *Transaction
}

// Next returns false if iterator is at its end. Otherwise, it returns true.
// Note: False does not stop the iterator. Release() should be called.
func (i *substateIterator) Next() bool {
	i.cur = <-i.resultCh
	return i.cur != nil
}

// Error returns iterators error if any.
func (i *substateIterator) Error() error {
	return errors.Join(i.err, i.iter.Error())
}

// Value returns current value hold by the iterator.
func (i *substateIterator) Value() *Transaction {
	return i.cur
}

// Release the iterator and wait until all threads are closed gracefully.
func (i *substateIterator) Release() {
	i.iter.Release()
	i.wg.Wait()
}

type Transaction struct {
	Block       uint64
	Transaction int
	Substate    *new_substate.Substate
}

type rawEntry struct {
	key   []byte
	value []byte
}

func (i *substateIterator) toTransaction(data rawEntry) (*Transaction, error) {
	key := data.key
	value := data.value

	block, tx, err := DecodeStage1SubstateKey(data.key)
	if err != nil {
		return nil, fmt.Errorf("invalid substate key: %v; %v", key, err)
	}

	rlpSubstate, err := rlp.Decode(value, block)
	if err != nil {
		return nil, err
	}

	ss, err := rlpSubstate.ToSubstate(i.db.GetCode)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Block:       block,
		Transaction: tx,
		Substate:    ss,
	}, nil
}
