package db

import (
	"encoding/binary"
	"fmt"

	gethrlp "github.com/Fantom-foundation/Substate/geth/rlp"
	"github.com/Fantom-foundation/Substate/new_substate"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// SubstateDB is a wrapper around CodeDB. It extends it with Has/Get/Put/DeleteSubstate functions.
type SubstateDB interface {
	CodeDB

	// HasSubstate returns true if the DB does contain Substate for given block and tx number.
	HasSubstate(block uint64, tx int) (bool, error)

	// GetSubstate gets the Substate for given block and tx number.
	GetSubstate(block uint64, tx int) (*new_substate.Substate, error)

	// PutSubstate inserts given substate to DB including the block and tx number.
	PutSubstate(block uint64, tx int, substate *new_substate.Substate) error

	// DeleteSubstate deletes Substate for given block and tx number.
	DeleteSubstate(block uint64, tx int) error

	NewSubstateIterator(start int, numWorkers int) SubstateIterator
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
	return db.Has(Stage1SubstateKey(block, tx))
}

// GetSubstate returns substate for given block and tx number if exists within DB.
func (db *substateDB) GetSubstate(block uint64, tx int) (*new_substate.Substate, error) {
	val, err := db.Get(Stage1SubstateKey(block, tx))
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

	return rlpSubstate.ToSubstate(db.GetCode)
}

func (db *substateDB) PutSubstate(block uint64, tx int, ss *new_substate.Substate) error {
	for i, account := range ss.InputAlloc {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put input-alloc code from substate-account %v block %v, %v tx into db; %v", i, block, tx, err)
		}
	}

	for i, account := range ss.OutputAlloc {
		err := db.PutCode(account.Code)
		if err != nil {
			return fmt.Errorf("cannot put ouput-alloc code from substate-account %v block %v, %v tx into db; %v", i, block, tx, err)
		}
	}

	if msg := ss.Message; msg.To == nil {
		err := db.PutCode(msg.Data)
		if err != nil {
			return fmt.Errorf("cannot put input data from substate block %v, %v tx into db; %v", block, tx, err)
		}
	}

	key := Stage1SubstateKey(block, tx)

	substateRLP := rlp.NewRLP(ss)
	value, err := gethrlp.EncodeToBytes(substateRLP)
	if err != nil {
		return fmt.Errorf("cannot encode substate-rlp block %v, tx %v; %v", block, tx, err)
	}

	return db.Put(key, value)
}

func (db *substateDB) DeleteSubstate(block uint64, tx int) error {
	return db.Delete(Stage1SubstateKey(block, tx))
}

// NewSubstateIterator returns iterator which iterates over Substates.
func (db *substateDB) NewSubstateIterator(start int, numWorkers int) SubstateIterator {
	num := make([]byte, 4)
	binary.BigEndian.PutUint32(num, uint32(start))
	iter := newSubstateIterator(db, num)

	// Create channels
	errCh := make(chan error)
	rawDataChs := make([]chan rawEntry, numWorkers)
	resultChs := make([]chan *Transaction, numWorkers)

	for i := 0; i < numWorkers; i++ {
		rawDataChs[i] = make(chan rawEntry, 10)
		resultChs[i] = make(chan *Transaction, 10)
	}

	// Start iter => raw data stage
	iter.wg.Add(1)
	go func() {
		defer func() {
			for _, c := range rawDataChs {
				close(c)
			}
			iter.wg.Done()
		}()
		step := 0
		for {
			if !iter.iter.Next() {
				return
			}

			key := make([]byte, len(iter.iter.Key()))
			copy(key, iter.iter.Key())
			value := make([]byte, len(iter.iter.Value()))
			copy(value, iter.iter.Value())

			res := rawEntry{key, value}

			select {
			case e := <-errCh:
				iter.err = e
				return
			case rawDataChs[step] <- res: // fall-through
			}
			step = (step + 1) % numWorkers
		}
	}()

	// Start raw data => parsed transaction stage (parallel)
	for i := 0; i < numWorkers; i++ {
		iter.wg.Add(1)
		id := i

		go func() {
			defer func() {
				close(resultChs[id])
				iter.wg.Done()
			}()
			for raw := range rawDataChs[id] {
				transaction, err := iter.toTransaction(raw)
				if err != nil {
					errCh <- err
					return
				}
				resultChs[id] <- transaction
			}
		}()
	}

	// Start the go routine moving transactions from parsers to sink in order
	iter.wg.Add(1)
	go func() {
		defer func() {
			close(iter.resultCh)
			iter.wg.Done()
		}()
		step := 0
		for openProducers := numWorkers; openProducers > 0; {
			next := <-resultChs[step%numWorkers]
			if next != nil {
				iter.resultCh <- next
			} else {
				openProducers--
			}
			step++
		}
	}()

	return iter
}

// BlockToBytes returns binary BigEndian representation of given block number.
func BlockToBytes(block uint64) []byte {
	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes[0:8], block)
	return blockBytes
}

// Stage1SubstateKey returns Stage1SubstatePrefix with appended
// block number creating key used in baseDB for Substates.
func Stage1SubstateKey(block uint64, tx int) []byte {
	prefix := []byte(Stage1SubstatePrefix)

	blockTx := make([]byte, 16)
	binary.BigEndian.PutUint64(blockTx[0:8], block)
	binary.BigEndian.PutUint64(blockTx[8:16], uint64(tx))

	return append(prefix, blockTx...)
}

// Stage1SubstateBlockPrefix returns Stage1SubstatePrefix with appended
// block number creating prefix used in baseDB for Substates.
func Stage1SubstateBlockPrefix(block uint64) []byte {
	return append([]byte(Stage1SubstatePrefix), BlockToBytes(block)...)
}

// DecodeStage1SubstateKey decodes key created by Stage1SubstateBlockPrefix back to block and tx number.
func DecodeStage1SubstateKey(key []byte) (block uint64, tx int, err error) {
	prefix := Stage1SubstatePrefix
	if len(key) != len(prefix)+8+8 {
		err = fmt.Errorf("invalid length of stage1 substate key: %v", len(key))
		return
	}
	if p := string(key[:len(prefix)]); p != prefix {
		err = fmt.Errorf("invalid prefix of stage1 substate key: %#x", p)
		return
	}
	blockTx := key[len(prefix):]
	block = binary.BigEndian.Uint64(blockTx[0:8])
	tx = int(binary.BigEndian.Uint64(blockTx[8:16]))
	return
}
