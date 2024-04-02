package substate

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/rawdb"
)

// NewDb creates a new instance of DB from given path.
func NewDb(path string, readOnly bool) (*DB, error) {
	backend, err := rawdb.NewLevelDBDatabase(path, 1024, 100, substateNamespace, readOnly)
	if err != nil {
		return nil, fmt.Errorf("error opening leveldb %s; %v", substateDir, err)
	}
	return newSubstateDB(backend), nil
}

// MakeDb creates a new instance of DB from given backend.
func MakeDb(backend BackendDatabase) *DB {
	// todo Rename to OpenSubstateDb after deprecated functions are removed
	return newSubstateDB(backend)
}

// NewInMemoryDb creates new instance of in-memory DB
func NewInMemoryDb() *DB {
	return newSubstateDB(rawdb.NewMemoryDatabase())
}
