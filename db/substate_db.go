package db

import (
	"encoding/binary"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/urfave/cli/v2"

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

	// GetBlockSubstates returns all existing substates for given block.
	GetBlockSubstates(block uint64) (map[int]*substate.Substate, error)

	// PutSubstate inserts given substate to DB.
	PutSubstate(substate *substate.Substate) error

	// DeleteSubstate deletes Substate for given block and tx number.
	DeleteSubstate(block uint64, tx int) error

	NewSubstateIterator(start int, numWorkers int) Iterator[*substate.Substate]

	NewSubstateTaskPool(name string, taskFunc SubstateTaskFunc, first, last uint64, ctx *cli.Context) *SubstateTaskPool

	// GetFirstSubstate returns last substate (block and transaction wise) inside given DB.
	GetFirstSubstate() *substate.Substate

	// GetLastSubstate returns last substate (block and transaction wise) inside given DB.
	GetLastSubstate() (*substate.Substate, error)
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

func MakeDefaultSubstateDB(db *leveldb.DB) SubstateDB {
	return &substateDB{&codeDB{&baseDB{backend: db}}}
}

func MakeDefaultSubstateDBFromBaseDB(db BaseDB) SubstateDB {
	return &substateDB{&codeDB{&baseDB{backend: db.getBackend()}}}
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

func (db *substateDB) GetFirstSubstate() *substate.Substate {
	iter := db.NewSubstateIterator(0, 1)

	defer iter.Release()

	if iter.Next() {
		return iter.Value()
	}

	return nil
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

	rlpSubstate, err := rlp.Decode(val)
	if err != nil {
		return nil, fmt.Errorf("cannot decode data into rlp block: %v, tx %v; %v", block, tx, err)
	}

	return rlpSubstate.ToSubstate(db.GetCode, block, tx)
}

// GetBlockSubstates returns substates for given block if exists within DB.
func (db *substateDB) GetBlockSubstates(block uint64) (map[int]*substate.Substate, error) {
	var err error

	txSubstate := make(map[int]*substate.Substate)

	prefix := SubstateDBBlockPrefix(block)

	iter := db.backend.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		b, tx, err := DecodeSubstateDBKey(key)
		if err != nil {
			return nil, fmt.Errorf("record-replay: invalid substate key found for block %v: %v", block, err)
		}

		if block != b {
			return nil, fmt.Errorf("record-replay: GetBlockSubstates(%v) iterated substates from block %v", block, b)
		}

		rlpSubstate, err := rlp.Decode(value)
		if err != nil {
			return nil, fmt.Errorf("cannot decode data into rlp block: %v, tx %v; %v", block, tx, err)
		}

		sbstt, err := rlpSubstate.ToSubstate(db.GetCode, block, tx)
		if err != nil {
			return nil, fmt.Errorf("cannot decode data into substate: %v", err)
		}

		txSubstate[tx] = sbstt
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, err
	}

	return txSubstate, nil
}

func (db *substateDB) PutSubstate(ss *substate.Substate) error {
	for i, account := range ss.InputSubstate {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put preState code from substate-account %v block %v, %v tx into db; %v", i, ss.Block, ss.Transaction, err)
		}
	}

	for i, account := range ss.OutputSubstate {
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
	blockTx := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTx, uint64(start))
	iter := newSubstateIterator(db, blockTx)

	iter.start(numWorkers)

	return iter
}

func (db *substateDB) NewSubstateTaskPool(name string, taskFunc SubstateTaskFunc, first, last uint64, ctx *cli.Context) *SubstateTaskPool {
	return &SubstateTaskPool{
		Name:     name,
		TaskFunc: taskFunc,

		First: first,
		Last:  last,

		Workers:         ctx.Int(WorkersFlag.Name),
		SkipTransferTxs: ctx.Bool(SkipTransferTxsFlag.Name),
		SkipCallTxs:     ctx.Bool(SkipCallTxsFlag.Name),
		SkipCreateTxs:   ctx.Bool(SkipCreateTxsFlag.Name),

		Ctx: ctx,

		DB: db,
	}
}

func (db *substateDB) GetLastSubstate() (*substate.Substate, error) {
	block, err := db.GetLastBlock()
	if err != nil {
		return nil, err
	}
	substates, err := db.GetBlockSubstates(block)
	if err != nil {
		return nil, fmt.Errorf("cannot get block substates; %w", err)
	}
	if len(substates) == 0 {
		return nil, fmt.Errorf("block %v doesn't have any substates.", block)
	}
	maxTx := 0
	for txIdx, _ := range substates {
		if txIdx > maxTx {
			maxTx = txIdx
		}
	}
	return substates[maxTx], nil
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
