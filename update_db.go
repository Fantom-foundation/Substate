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
	SubstateAllocPrefix     = "2s" // SubstateAllocPrefix + block (64-bit) + tx (64-bit) -> substateRLP
	SubstateAllocCodePrefix = "2c" // SubstateAllocCodePrefix + codeHash (256-bit) -> code
)

func SubstateAllocKey(block uint64) []byte {
	prefix := []byte(SubstateAllocPrefix)
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	return append(prefix, blockTx...)
}

func DecodeSubstateAllocKey(key []byte) (block uint64, err error) {
	prefix := SubstateAllocPrefix
	if len(key) != len(prefix)+8 {
		err = fmt.Errorf("invalid length of stage1 substate key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of stage1 substate key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	return
}

func SubstateAllocBlockPrefix(block uint64) []byte {
	prefix := []byte(SubstateAllocPrefix)

	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes[0:8], block)

	return append(prefix, blockBytes...)
}

type UpdateDB struct {
	backend BackendDatabase
}

func NewUpdateDB(backend BackendDatabase) *UpdateDB {
	return &UpdateDB{backend: backend}
}

func OpenUpdateDB(updateSetDir string) *UpdateDB {
	fmt.Println("record-replay: OpenUpdateSetDB")
	backend, err := rawdb.NewLevelDBDatabase(updateSetDir, 1024, 100, "updatesetdir", false)
	if err != nil {
		panic(fmt.Errorf("error opening update-set leveldb %s: %v", updateSetDir, err))
	}
	return NewUpdateDB(backend)
}

func OpenUpdateDBReadOnly(updateSetDir string) *UpdateDB {
	fmt.Println("record-replay: OpenUpdateSetDB")
	backend, err := rawdb.NewLevelDBDatabase(updateSetDir, 1024, 100, "updatesetdir", true)
	if err != nil {
		panic(fmt.Errorf("error opening update-set leveldb %s: %v", updateSetDir, err))
	}
	return NewUpdateDB(backend)
}

func (db *UpdateDB) Compact(start []byte, limit []byte) error {
	return db.backend.Compact(start, limit)
}

func (db *UpdateDB) Close() error {
	return db.backend.Close()
}

func (db *UpdateDB) HasCode(codeHash common.Hash) bool {
	if codeHash == EmptyCodeHash {
		return false
	}
	key := Stage1CodeKey(codeHash)
	has, err := db.backend.Has(key)
	if err != nil {
		panic(fmt.Errorf("record-replay: error checking bytecode for codeHash %s: %v", codeHash.Hex(), err))
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
		panic(fmt.Errorf("record-replay: error getting code %s: %v", codeHash.Hex(), err))
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
		panic(fmt.Errorf("record-replay: error putting code %s: %v", codeHash.Hex(), err))
	}
}

func (db *UpdateDB) HasUpdateSet(block uint64) bool {
	key := SubstateAllocKey(block)
	has, _ := db.backend.Has(key)
	return has
}

func (alloc *SubstateAlloc) SetUpdateSetRLP(allocRLP SubstateAllocRLP, db *UpdateDB) {
	*alloc = make(SubstateAlloc)
	for i, addr := range allocRLP.Addresses {
		var sa SubstateAccount

		saRLP := allocRLP.Accounts[i]
		sa.Balance = saRLP.Balance
		sa.Nonce = saRLP.Nonce
		sa.Code = db.GetCode(saRLP.CodeHash)
		sa.Storage = make(map[common.Hash]common.Hash)
		for i := range saRLP.Storage {
			sa.Storage[saRLP.Storage[i][0]] = saRLP.Storage[i][1]
		}

		(*alloc)[addr] = &sa
	}
}

func (db *UpdateDB) GetUpdateSet(block uint64) *SubstateAlloc {
	var err error
	key := SubstateAllocKey(block)
	value, err := db.backend.Get(key)
	if err != nil {
		panic(fmt.Errorf("record-replay: error getting substate %v from substate DB: %v,", block, err))
	}
	// try decoding as substates from latest hard forks
	updateSetRLP := UpdateSetRLP{}
	err = rlp.DecodeBytes(value, &updateSetRLP)
	updateSet := SubstateAlloc{}
	updateSet.SetUpdateSetRLP(updateSetRLP.SubstateAlloc, db)
	return &updateSet
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
			panic(fmt.Errorf("record-replay: error putting update-set %v into substate DB: %v", block, err))
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

	block, err := DecodeSubstateAllocKey(data.key)
	if err != nil {
		panic(fmt.Errorf("record-replay: invalid update-set key found: %v - issue: %v", key, err))
	}

	updateSetRLP := UpdateSetRLP{}
	rlp.DecodeBytes(value, &updateSetRLP)
	updateSet := SubstateAlloc{}
	updateSet.SetUpdateSetRLP(updateSetRLP.SubstateAlloc, db)

	return &UpdateBlock{
		Block:           block,
		UpdateSet:       &updateSet,
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

func NewUpdateSetIterator(db *UpdateDB, startBlock, endBlock uint64, workers int) UpdateSetIterator {
	start := SubstateAllocBlockPrefix(startBlock)
	iter := db.backend.NewIterator(nil, start)

	// Create channels
	done := make(chan int)
	rawData := make([]chan rawEntry, workers)
	results := make([]chan *UpdateBlock, workers)
	result := make(chan *UpdateBlock, 10)

	for i := 0; i < workers; i++ {
		rawData[i] = make(chan rawEntry, 1)
		results[i] = make(chan *UpdateBlock, 1)
	}

	// Start iter => raw data stage
	go func() {
		defer func() {
			for _, c := range rawData {
				close(c)
			}
		}()
		step := 0
		for {
			if !iter.Next() {
				return
			}

			key := make([]byte, len(iter.Key()))
			copy(key, iter.Key())

			// Decode key, if past the end block, stop here.
			// This avoids filling channels which huge data objects that are not consumed.
			block, err := DecodeSubstateAllocKey(key)
			if err != nil {
				panic(fmt.Errorf("worldstate-upate: invalid update-set key found: %v - issue: %v", key, err))
			}
			if block > endBlock {
				return
			}

			value := make([]byte, len(iter.Value()))
			copy(value, iter.Value())

			res := rawEntry{key, value}

			select {
			case <-done:
				return
			case rawData[step] <- res: // fall-through
			}
			step = (step + 1) % workers
		}
	}()

	// Start raw data => parsed transaction stage (parallel)
	for i := 0; i < workers; i++ {
		id := i
		go func() {
			defer close(results[id])
			for raw := range rawData[id] {
				results[id] <- parseUpdateSet(db, raw)
			}
		}()
	}

	// Start the go routine moving transactions from parsers to sink in order
	go func() {
		defer close(result)
		step := 0
		for openProducers := workers; openProducers > 0; {
			next := <-results[step%workers]
			if next != nil {
				result <- next
			} else {
				openProducers--
			}
			step++
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
