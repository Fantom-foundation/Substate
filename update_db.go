package substate

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

type UpdateSetRLP struct {
	SubstateAlloc   SubstateAllocRLP
	DeletedAccounts []common.Address
}

func NewUpdateSetRLP(updateset SubstateAlloc, deletedAccounts []common.Address) UpdateSetRLP {
	var rlp UpdateSetRLP

	rlp.SubstateAlloc = NewSubstateAllocRLP(updateset)
	rlp.DeletedAccounts = deletedAccounts
	return rlp

}

const (
	SubstateAllocPrefix = "2s" // SubstateAllocPrefix + block (64-bit) + tx (64-bit) -> substateRLP
)

func SubstateAllocKey(block uint64) []byte {
	prefix := []byte(SubstateAllocPrefix)
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	return append(prefix, blockTx...)
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

type UpdateDB struct {
	backend BackendDatabase
}

func NewUpdateDB(backend BackendDatabase) *UpdateDB {
	return &UpdateDB{backend: backend}
}

func OpenUpdateDB(updateSetDir string) (*UpdateDB, error) {
	fmt.Println("substate: OpenUpdateSetDB")
	backend, err := rawdb.NewLevelDBDatabase(updateSetDir, 1024, 100, "updatesetdir", false)
	if err != nil {
		return nil, fmt.Errorf("error opening update-set leveldb %s: %v", updateSetDir, err)
	}
	return NewUpdateDB(backend), nil
}

func OpenUpdateDBReadOnly(updateSetDir string) (*UpdateDB, error) {
	fmt.Println("substate: OpenUpdateSetDB")
	backend, err := rawdb.NewLevelDBDatabase(updateSetDir, 1024, 100, "updatesetdir", true)
	if err != nil {
		return nil, fmt.Errorf("error opening update-set leveldb %s: %v", updateSetDir, err)
	}
	return NewUpdateDB(backend), nil
}

func (db *UpdateDB) Compact(start []byte, limit []byte) error {
	return db.backend.Compact(start, limit)
}

func (db *UpdateDB) Close() error {
	return db.backend.Close()
}

func (db *UpdateDB) GetLastKey() uint64 {
	var block uint64
	var err error
	iter := db.backend.NewIterator([]byte(SubstateAllocPrefix), nil)
	for iter.Next() {
		block, err = DecodeUpdateSetKey(iter.Key())
		if err != nil {
			panic(fmt.Errorf("error iterating updateDB: %v", err))
		}
	}
	iter.Release()
	return block
}

func (db *UpdateDB) HasCode(codeHash common.Hash) bool {
	if codeHash == EmptyCodeHash {
		return false
	}
	key := Stage1CodeKey(codeHash)
	has, err := db.backend.Has(key)
	if err != nil {
		panic(fmt.Errorf("substate: error checking bytecode for codeHash %s: %v", codeHash.Hex(), err))
	}
	return has
}

func (db *UpdateDB) GetCode(codeHash common.Hash) []byte {
	if codeHash == EmptyCodeHash {
		return nil
	}
	key := Stage1CodeKey(codeHash)
	code, err := db.backend.Get(key)
	if err != nil {
		panic(fmt.Errorf("substate: error getting code %s: %v", codeHash.Hex(), err))
	}
	return code
}

func (db *UpdateDB) PutCode(code []byte) {
	if len(code) == 0 {
		return
	}
	codeHash := crypto.Keccak256Hash(code)
	key := Stage1CodeKey(codeHash)
	err := db.backend.Put(key, code)
	if err != nil {
		panic(fmt.Errorf("substate: error putting code %s: %v", codeHash.Hex(), err))
	}
}

func (db *UpdateDB) HasUpdateSet(block uint64) bool {
	key := SubstateAllocKey(block)
	has, _ := db.backend.Has(key)
	return has
}

func (up *UpdateSetRLP) GetSubstateAlloc(db *UpdateDB) *SubstateAlloc {
	alloc := make(SubstateAlloc)
	for i, addr := range up.SubstateAlloc.Addresses {
		var sa SubstateAccount
		saRLP := up.SubstateAlloc.Accounts[i]
		sa.Balance = saRLP.Balance
		sa.Nonce = saRLP.Nonce
		sa.Code = db.GetCode(saRLP.CodeHash)
		sa.Storage = make(map[common.Hash]common.Hash)
		for i := range saRLP.Storage {
			sa.Storage[saRLP.Storage[i][0]] = saRLP.Storage[i][1]
		}
		alloc[addr] = &sa
	}
	return &alloc
}

func (db *UpdateDB) GetUpdateSet(block uint64) *SubstateAlloc {
	var err error
	key := SubstateAllocKey(block)
	value, err := db.backend.Get(key)
	if err != nil {
		panic(fmt.Errorf("substate: error getting substate %v from substate DB: %v,", block, err))
	}
	// decode value
	updateSetRLP := UpdateSetRLP{}
	if err := rlp.DecodeBytes(value, &updateSetRLP); err != nil {
		panic(fmt.Errorf("substate: failed to decode updateset value at block %v, key %v", block, key))
	}
	updateSet := updateSetRLP.GetSubstateAlloc(db)
	return updateSet
}

func (db *UpdateDB) PutUpdateSet(block uint64, updateSet *SubstateAlloc, deletedAccounts []common.Address) {
	var err error

	// put deployed/creation code
	for _, account := range *updateSet {
		db.PutCode(account.Code)
	}
	key := SubstateAllocKey(block)
	defer func() {
		if err != nil {
			panic(fmt.Errorf("substate: error putting update-set %v into substate DB: %v", block, err))
		}
	}()

	updateSetRLP := NewUpdateSetRLP(*updateSet, deletedAccounts)

	value, err := rlp.EncodeToBytes(updateSetRLP)
	if err != nil {
		panic(err)
	}
	err = db.backend.Put(key, value)
	if err != nil {
		panic(err)
	}
}

func (db *UpdateDB) DeleteSubstateAlloc(block uint64) {
	key := SubstateAllocKey(block)
	err := db.backend.Delete(key)
	if err != nil {
		panic(err)
	}
}

type UpdateBlock struct {
	Block           uint64
	UpdateSet       *SubstateAlloc
	DeletedAccounts []common.Address
}

func parseUpdateSet(db *UpdateDB, data rawEntry) *UpdateBlock {
	key := data.key
	value := data.value

	block, err := DecodeUpdateSetKey(data.key)
	if err != nil {
		panic(fmt.Errorf("substate: invalid update-set key found: %v - issue: %v", key, err))
	}

	updateSetRLP := UpdateSetRLP{}
	rlp.DecodeBytes(value, &updateSetRLP)
	updateSet := updateSetRLP.GetSubstateAlloc(db)

	return &UpdateBlock{
		Block:           block,
		UpdateSet:       updateSet,
		DeletedAccounts: updateSetRLP.DeletedAccounts,
	}
}

type UpdateSetIterator struct {
	db   *UpdateDB
	iter ethdb.Iterator
	cur  *UpdateBlock

	// Connections to parsing pipeline
	source <-chan *UpdateBlock
	done   chan<- int
}

func NewUpdateSetIterator(db *UpdateDB, startBlock, endBlock uint64) UpdateSetIterator {
	start := BlockToBytes(startBlock)
	// updateset prefix is already in start
	iter := db.backend.NewIterator([]byte(SubstateAllocPrefix), start)

	done := make(chan int)
	result := make(chan *UpdateBlock, 1)

	go func() {
		defer close(result)
		for iter.Next() {

			key := make([]byte, len(iter.Key()))
			copy(key, iter.Key())

			// Decode key, if past the end block, stop here.
			// This avoids filling channels which huge data objects that are not consumed.
			block, err := DecodeUpdateSetKey(key)
			if err != nil {
				panic(fmt.Errorf("worldstate-upate: invalid update-set key found: %v - issue: %v", key, err))
			}
			if block > endBlock {
				return
			}

			value := make([]byte, len(iter.Value()))
			copy(value, iter.Value())

			raw := rawEntry{key, value}

			select {
			case <-done:
				return
			case result <- parseUpdateSet(db, raw): //fall-through
			}
		}
	}()

	return UpdateSetIterator{
		db:     db,
		iter:   iter,
		source: result,
		done:   done,
	}
}

func (i *UpdateSetIterator) Release() {
	close(i.done)

	// drain pipeline until the result channel is closed
	for open := true; open; _, open = <-i.source {
	}

	i.iter.Release()
}

func (i *UpdateSetIterator) Next() bool {
	if i.iter == nil {
		return false
	}
	i.cur = <-i.source
	return i.cur != nil
}

func (i *UpdateSetIterator) Value() *UpdateBlock {
	return i.cur
}
