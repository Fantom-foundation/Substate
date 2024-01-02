package db

import (
	"errors"
	"fmt"
	"io"

	"github.com/syndtr/goleveldb/leveldb"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	Stage1SubstatePrefix = "1s" // Stage1SubstatePrefix + block (64-bit) + tx (64-bit) -> substateRLP
	Stage1CodePrefix     = "1c" // Stage1CodePrefix + codeHash (256-bit) -> code
)

// KeyValueWriter wraps the Put method of a backing data store.
type KeyValueWriter interface {
	// Put inserts the given value into the key-value data store.
	Put(key []byte, value []byte) error

	// Delete removes the key from the key-value data store.
	Delete(key []byte) error
}

type BaseDB interface {
	KeyValueWriter

	io.Closer

	// Has returns true if the baseDB does contain the given key.
	Has([]byte) (bool, error)

	// Get gets the value for the given key.
	Get([]byte) ([]byte, error)

	// NewBatch creates a write-only database that buffers changes to its host db
	// until a final write is called.
	NewBatch() Batch

	// newIterator creates a binary-alphabetical iterator over a subset
	// of database content with a particular key prefix, starting at a particular
	// initial key (or after, if it does not exist).
	//
	// Note: This method assumes that the prefix is NOT part of the start, so there's
	// no need for the caller to prepend the prefix to the start
	NewIterator(prefix []byte, start []byte) ldbiterator.Iterator

	// Stat returns a particular internal stat of the database.
	Stat(property string) (string, error)

	// Compact flattens the underlying data store for the given key range. In essence,
	// deleted and overwritten versions are discarded, and the data is rearranged to
	// reduce the cost of operations needed to access them.
	//
	// A nil start is treated as a key before all keys in the data store; a nil limit
	// is treated as a key after all keys in the data store. If both is nil then it
	// will compact entire data store.
	Compact(start []byte, limit []byte) error

	// Close closes the DB. This will also release any outstanding snapshot,
	// abort any in-flight compaction and discard open transaction.
	//
	// Note:
	// It is not safe to close a DB until all outstanding iterators are released.
	// It is valid to call Close multiple times.
	// Other methods should not be called after the DB has been closed.
	Close() error
}

// NewDefaultBaseDB creates new instance of BaseDB with default options.
func NewDefaultBaseDB(path string) (BaseDB, error) {
	return newBaseDB(path, nil, nil, nil)
}

// NewBaseDB creates new instance of BaseDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewBaseDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (BaseDB, error) {
	return newBaseDB(path, o, wo, ro)
}

func newBaseDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*baseDB, error) {
	b, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, fmt.Errorf("cannot open leveldb; %v", err)
	}
	return &baseDB{
		backend: b,
		wo:      wo,
		ro:      ro,
	}, nil
}

// baseDB implements method needed by all three types of DBs.
type baseDB struct {
	backend *leveldb.DB
	wo      *opt.WriteOptions
	ro      *opt.ReadOptions
}

func (db *baseDB) Put(key []byte, value []byte) error {
	return db.backend.Put(key, value, db.wo)
}

func (db *baseDB) Delete(key []byte) error {
	return db.backend.Delete(key, db.wo)
}

func (db *baseDB) Close() error {
	return db.backend.Close()
}

func (db *baseDB) Has(key []byte) (bool, error) {
	return db.backend.Has(key, db.ro)
}

func (db *baseDB) Get(key []byte) ([]byte, error) {
	b, err := db.backend.Get(key, db.ro)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, nil
		}
	}
	return b, nil
}

func (db *baseDB) NewBatch() Batch {
	return newBatch(db.backend)
}

// newIterator returns iterator which iterates over values depending on the prefix.
// Note: If prefix is nil, everything is iterated.
func (db *baseDB) NewIterator(prefix []byte, start []byte) ldbiterator.Iterator {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return db.backend.NewIterator(r, db.ro)
}

func (db *baseDB) Stat(property string) (string, error) {
	return db.backend.GetProperty(property)
}

func (db *baseDB) Compact(start []byte, limit []byte) error {
	return db.backend.CompactRange(util.Range{Start: start, Limit: limit})
}
