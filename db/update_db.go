package db

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Fantom-foundation/Substate/geth/common"
	gethrlp "github.com/Fantom-foundation/Substate/geth/rlp"
	"github.com/Fantom-foundation/Substate/update_set"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	SubstateAllocPrefix = "2s" // SubstateAllocPrefix + block (64-bit) + tx (64-bit) -> substateRLP
)

// UpdateDB represents a CodeDB with in which the UpdateSet is inserted.
type UpdateDB interface {
	CodeDB

	// GetFirstKey returns block number of first UpdateSet. It returns an error if no UpdateSet is found.
	GetFirstKey() (uint64, error)

	// GetLastKey returns block number of last UpdateSet. It returns an error if no UpdateSet is found.
	GetLastKey() (uint64, error)

	// HasUpdateSet returns true if there is an UpdateSet on given block.
	HasUpdateSet(block uint64) (bool, error)

	// GetUpdateSet returns UpdateSet for given block. If there is not an UpdateSet for the block, nil is returned.
	GetUpdateSet(block uint64) (*update_set.UpdateSet, error)

	// PutUpdateSet inserts the UpdateSet with deleted accounts into the DB assigned to given block.
	PutUpdateSet(updateSet *update_set.UpdateSet, deletedAccounts []common.Address) error

	// DeleteUpdateSet deletes UpdateSet for given block. It returns an error if there is no UpdateSet on given block.
	DeleteUpdateSet(block uint64) error

	NewUpdateSetIterator(start, end uint64) Iterator[*update_set.UpdateSet]
}

// NewDefaultUpdateDB creates new instance of UpdateDB with default options.
func NewDefaultUpdateDB(path string) (UpdateDB, error) {
	return newUpdateDB(path, nil, nil, nil)
}

// NewUpdateDB creates new instance of UpdateDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewUpdateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (UpdateDB, error) {
	return newUpdateDB(path, o, wo, ro)
}

func newUpdateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*updateDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &updateDB{base}, nil
}

type updateDB struct {
	*codeDB
}

func (db *updateDB) GetFirstKey() (uint64, error) {
	r := util.BytesPrefix([]byte(SubstateAllocPrefix))

	iter := db.backend.NewIterator(r, db.ro)
	defer iter.Release()

	for iter.Next() {
		firstBlock, err := DecodeUpdateSetKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}
		return firstBlock, nil
	}
	return 0, errors.New("no updateset found")
}

func (db *updateDB) GetLastKey() (uint64, error) {
	r := util.BytesPrefix([]byte(SubstateAllocPrefix))

	iter := db.backend.NewIterator(r, nil)
	defer iter.Release()

	for iter.Next() {
		lastBlock, err := DecodeUpdateSetKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}

		return lastBlock, nil
	}

	return 0, errors.New("no updateset found")
}

func (db *updateDB) HasUpdateSet(block uint64) (bool, error) {
	key := SubstateAllocKey(block)
	return db.Has(key)
}

func (db *updateDB) GetUpdateSet(block uint64) (*update_set.UpdateSet, error) {
	key := SubstateAllocKey(block)
	value, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get updateset block: %v, key %v; %v", block, key, err)
	}

	if value == nil {
		return nil, nil
	}

	// decode value
	var updateSetRLP update_set.UpdateSetRLP
	if err = gethrlp.DecodeBytes(value, &updateSetRLP); err != nil {
		return nil, fmt.Errorf("cannot decode update-set rlp block: %v, key %v; %v", block, key, err)
	}

	return updateSetRLP.ToSubstateAlloc(db.GetCode, block)
}

func (db *updateDB) PutUpdateSet(updateSet *update_set.UpdateSet, deletedAccounts []common.Address) error {
	// put deployed/creation code
	for _, account := range updateSet.Alloc {
		err := db.PutCode(account.Code)
		if err != nil {
			return err
		}
	}

	key := SubstateAllocKey(updateSet.Block)
	updateSetRLP := update_set.NewUpdateSetRLP(updateSet, deletedAccounts)

	value, err := gethrlp.EncodeToBytes(updateSetRLP)
	if err != nil {
		return fmt.Errorf("cannot encode update-set; %v", err)
	}

	return db.Put(key, value)
}

func (db *updateDB) DeleteUpdateSet(block uint64) error {
	key := SubstateAllocKey(block)
	return db.Delete(key)
}

func (db *updateDB) NewUpdateSetIterator(start, end uint64) Iterator[*update_set.UpdateSet] {
	iter := newUpdateSetIterator(db, start, end)

	iter.start(0)

	return iter
}

func DecodeUpdateSetKey(key []byte) (block uint64, err error) {
	prefix := SubstateAllocPrefix
	if len(key) != len(prefix)+8 {
		err = fmt.Errorf("invalid length of updateset key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of updateset key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	return
}

func SubstateAllocKey(block uint64) []byte {
	prefix := []byte(SubstateAllocPrefix)
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	return append(prefix, blockTx...)
}
