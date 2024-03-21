package db

import (
	"encoding/binary"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/Fantom-foundation/Substate/substate"
	trlp "github.com/Fantom-foundation/Substate/types/rlp"
)

const SubstateDBPrefix = "1s" // SubstateDBPrefix + block (64-bit) + tx (64-bit) -> substateRLP

// SubstateDB is a wrapper around CodeDB. It extends it with Has/Get/Put/DeleteSubstate functions.
type SubstateDB interface {
	CodeDB

	// HasSubstate returns true if the DB does contain Substate for given block and tx number.
	HasSubstate(block uint64, tx int) (bool, error)

	// GetSubstate gets the Substate for given block and tx number.
	GetSubstate(block uint64, tx int) (*substate.Substate, error)

	// PutSubstate inserts given substate to DB.
	PutSubstate(substate *substate.Substate) error

	// DeleteSubstate deletes Substate for given block and tx number.
	DeleteSubstate(block uint64, tx int) error

	NewSubstateIterator(start int, numWorkers int) Iterator[*substate.Substate]
}

// NewDefaultSubstateDB creates new instance of SubstateDB with default options.
func NewDefaultSubstateDB(path string) (SubstateDB, error) {
	return newSubstateDB(path, nil, nil, nil)
}

// NewSubstateDB creates new instance of SubstateDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewSubstateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (SubstateDB, error) {
	return newSubstateDB(path, o, wo, ro)
}

func MakeDefaultSubstateDb(db *leveldb.DB) SubstateDB {
	return &substateDB{&codeDB{&baseDB{backend: db}}}
}

func MakeSubstateDb(db *leveldb.DB, wo *opt.WriteOptions, ro *opt.ReadOptions) SubstateDB {
	return &substateDB{&codeDB{&baseDB{backend: db, wo: wo, ro: ro}}}
}

func newSubstateDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*substateDB, error) {
	base, err := newCodeDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &substateDB{base}, nil
}

type substateDB struct {
	*codeDB
}

func (db *substateDB) HasSubstate(block uint64, tx int) (bool, error) {
	return db.Has(SubstateDBKey(block, tx))
}

// GetSubstate returns substate for given block and tx number if exists within DB.
func (db *substateDB) GetSubstate(block uint64, tx int) (*substate.Substate, error) {
	val, err := db.Get(SubstateDBKey(block, tx))
	if err != nil {
		return nil, fmt.Errorf("cannot get substate block: %v, tx: %v from db; %v", block, tx, err)
	}

	// not in db
	if val == nil {
		return nil, nil
	}

	rlpSubstate, err := rlp.Decode(val, block)
	if err != nil {
		return nil, fmt.Errorf("cannot decode data into rlp block: %v, tx %v; %v", block, tx, err)
	}

	return rlpSubstate.ToSubstate(db.GetCode, block, tx)
}

func (db *substateDB) PutSubstate(ss *substate.Substate) error {
	for i, account := range ss.PreState {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put preState code from substate-account %v block %v, %v tx into db; %v", i, ss.Block, ss.Transaction, err)
		}
	}

	for i, account := range ss.PostState {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put postState code from substate-account %v block %v, %v tx into db; %v", i, ss.Block, ss.Transaction, err)
		}
	}

	if msg := ss.Message; msg.To == nil {
		err := db.PutCode(msg.Data)
		if err != nil {
			return fmt.Errorf("cannot put input data from substate block %v, %v tx into db; %v", ss.Block, ss.Transaction, err)
		}
	}

	key := SubstateDBKey(ss.Block, ss.Transaction)

	substateRLP := rlp.NewRLP(ss)
	value, err := trlp.EncodeToBytes(substateRLP)
	if err != nil {
		return fmt.Errorf("cannot encode substate-rlp block %v, tx %v; %v", ss.Block, ss.Transaction, err)
	}

	return db.Put(key, value)
}

func (db *substateDB) DeleteSubstate(block uint64, tx int) error {
	return db.Delete(SubstateDBKey(block, tx))
}

// NewSubstateIterator returns iterator which iterates over Substates.
func (db *substateDB) NewSubstateIterator(start int, numWorkers int) Iterator[*substate.Substate] {
	num := make([]byte, 4)
	binary.BigEndian.PutUint32(num, uint32(start))
	iter := newSubstateIterator(db, num)

	iter.start(numWorkers)

	return iter
}

// BlockToBytes returns binary BigEndian representation of given block number.
func BlockToBytes(block uint64) []byte {
	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes[0:8], block)
	return blockBytes
}

// SubstateDBKey returns SubstateDBPrefix with appended
// block number creating key used in baseDB for Substates.
func SubstateDBKey(block uint64, tx int) []byte {
	prefix := []byte(SubstateDBPrefix)

	blockTx := make([]byte, 16)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	binary.BigEndian.PutUint64(blockTx[8:16], uint64(tx))

	return append(prefix, blockTx...)
}

// SubstateDBBlockPrefix returns SubstateDBPrefix with appended
// block number creating prefix used in baseDB for Substates.
func SubstateDBBlockPrefix(block uint64) []byte {
	return append([]byte(SubstateDBPrefix), BlockToBytes(block)...)
}

// DecodeSubstateDBKey decodes key created by SubstateDBBlockPrefix back to block and tx number.
func DecodeSubstateDBKey(key []byte) (block uint64, tx int, err error) {
	prefix := SubstateDBPrefix
	if len(key) != len(prefix)+8+8 {
		err = fmt.Errorf("invalid length of substate db key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of substate db key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	tx = int(binary.BigEndian.Uint64(blockTx[8:16]))
	return
}
