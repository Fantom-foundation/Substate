package db

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Fantom-foundation/Substate/geth/common"
	gethrlp "github.com/Fantom-foundation/Substate/geth/rlp"
	"github.com/Fantom-foundation/Substate/new_substate"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	SubstateAllocPrefix = "2s" // SubstateAllocPrefix + block (64-bit) + tx (64-bit) -> substateRLP
)

type UpdateDB interface {
	CodeDB

	GetFirstKey() (uint64, error)

	GetLastKey() (uint64, error)

	HasUpdateSet(block uint64) (bool, error)

	GetUpdateSet(block uint64) (new_substate.Alloc, error)

	PutUpdateSet(block uint64, updateSet new_substate.Alloc, deletedAccounts []common.Address) error

	DeleteUpdateSet(block uint64) error
}

type UpdateSetRLP struct {
	Alloc           rlp.Alloc
	DeletedAccounts []common.Address
}

func NewUpdateSetRLP(updateSet new_substate.Alloc, deletedAccounts []common.Address) UpdateSetRLP {
	return UpdateSetRLP{
		Alloc:           rlp.NewAlloc(updateSet),
		DeletedAccounts: deletedAccounts,
	}
}

func (up UpdateSetRLP) toSubstateAlloc(db *updateDB) (new_substate.Alloc, error) {
	alloc := make(new_substate.Alloc)

	for i, addr := range up.Alloc.Addresses {
		allocAcc := up.Alloc.Accounts[i]

		code, err := db.GetCode(allocAcc.CodeHash)
		if err != nil {
			return nil, err
		}

		acc := new_substate.Account{
			Nonce:   allocAcc.Nonce,
			Balance: allocAcc.Balance,
			Storage: make(map[common.Hash]common.Hash),
			Code:    code,
		}

		for j := range allocAcc.Storage {
			acc.Storage[up.Alloc.Accounts[j].Storage[j][0]] = up.Alloc.Accounts[j].Storage[j][1]
		}
		alloc[addr] = &acc
	}

	return alloc, nil
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

// GetFirstKey returns block number of first UpdateSet. It returns an error if no UpdateSet is found.
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

func (db *updateDB) GetUpdateSet(block uint64) (new_substate.Alloc, error) {
	key := SubstateAllocKey(block)
	value, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get updateset block: %v, key %v; %v", block, key, err)
	}

	if value == nil {
		return nil, nil
	}

	// decode value
	var updateSetRLP UpdateSetRLP
	if err = gethrlp.DecodeBytes(value, &updateSetRLP); err != nil {
		return nil, fmt.Errorf("cannot decode update-set rlp block: %v, key %v; %v", block, key, err)
	}

	return updateSetRLP.toSubstateAlloc(db)
}

func (db *updateDB) PutUpdateSet(block uint64, updateSet new_substate.Alloc, deletedAccounts []common.Address) error {
	// put deployed/creation code
	for _, account := range updateSet {
		err := db.PutCode(account.Code)
		if err != nil {
			return err
		}
	}

	key := SubstateAllocKey(block)
	updateSetRLP := NewUpdateSetRLP(updateSet, deletedAccounts)

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
