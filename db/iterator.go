package db

import (
	"errors"
	"sync"

	iter "github.com/Fantom-foundation/Substate/iterator"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
)

// Iterator iterates over a database's key/value pairs in ascending key order.
//
// When it encounters an error any seek will return false and will yield no key/
// value pairs. The error can be queried by calling the Error method. Calling
// Release is still necessary.
//
// An iterator must be released after use, but it is not necessary to read an
// iterator until exhaustion. An iterator is not safe for concurrent use, but it
// is safe to use multiple iterators concurrently.
type Iterator[T comparable] interface {
	iter.Iterator[T]

	// decode data returned from DB to given type T.
	decode(data rawEntry) (T, error)
}

type rawEntry struct {
	key   []byte
	value []byte
}

type iterator[T comparable] struct {
	err      error
	iter     ldbiterator.Iterator
	resultCh chan T
	wg       *sync.WaitGroup
	cur      T
}

func newIterator[T comparable](iter ldbiterator.Iterator) iterator[T] {
	return iterator[T]{
		iter:     iter,
		resultCh: make(chan T, 10),
		wg:       new(sync.WaitGroup),
	}
}

// Next returns false if iterator is at its end. Otherwise, it returns true.
// Note: False does not stop the iterator. Release() should be called.
func (i *iterator[T]) Next() bool {
	i.cur = <-i.resultCh
	var zero T
	return i.cur != zero
}

// Error returns iterators error if any.
func (i *iterator[T]) Error() error {
	return errors.Join(i.err, i.iter.Error())
}

// Value returns current value hold by the iterator.
func (i *iterator[T]) Value() T {
	return i.cur
}

// Release the iterator and wait until all threads are closed gracefully.
func (i *iterator[T]) Release() {
	i.iter.Release()
	i.wg.Wait()
}

func isNil[T comparable](arg T) bool {
	var t T
	return arg == t
}
