package db

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Batch interface {
	KeyValueWriter

	// ValueSize retrieves the amount of data queued up for writing.
	ValueSize() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch for reuse.
	Reset()

	// Replay replays the batch contents.
	Replay(w KeyValueWriter) error
}

// replayer is a small wrapper to implement the correct replay methods.
type replayer struct {
	writer  KeyValueWriter
	failure error
}

// Put inserts the given value into the key-value data store.
func (r *replayer) Put(key, value []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Put(key, value)
}

// Delete removes the key from the key-value data store.
func (r *replayer) Delete(key []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Delete(key)
}

func newBatch(db *leveldb.DB) Batch {
	return &batch{
		db: db,
		b:  new(leveldb.Batch),
	}
}

type batch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (b *batch) Put(key []byte, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

func (b *batch) ValueSize() int {
	return b.size
}

func (b *batch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

func (b *batch) Replay(w KeyValueWriter) error {
	return b.b.Replay(&replayer{writer: w})
}
