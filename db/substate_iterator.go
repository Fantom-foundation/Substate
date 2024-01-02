package db

import (
	"fmt"

	"github.com/Fantom-foundation/Substate/new_substate"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func newSubstateIterator(db *substateDB, start []byte) *substateIterator {
	r := util.BytesPrefix([]byte(Stage1SubstatePrefix))
	r.Start = append(r.Start, start...)

	return &substateIterator{
		iterator: newIterator[*new_substate.Substate](db.backend.NewIterator(r, db.ro)),
		db:       db,
	}
}

type substateIterator struct {
	iterator[*new_substate.Substate]
	db *substateDB
}

func (i *substateIterator) decode(data rawEntry) (*new_substate.Substate, error) {
	key := data.key
	value := data.value

	block, tx, err := DecodeStage1SubstateKey(data.key)
	if err != nil {
		return nil, fmt.Errorf("invalid substate key: %v; %v", key, err)
	}

	rlpSubstate, err := rlp.Decode(value, block)
	if err != nil {
		return nil, err
	}

	return rlpSubstate.ToSubstate(i.db.GetCode, block, tx)
}

func (i *substateIterator) start(numWorkers int) {
	// Create channels
	errCh := make(chan error)
	rawDataChs := make([]chan rawEntry, numWorkers)
	resultChs := make([]chan *new_substate.Substate, numWorkers)

	for i := 0; i < numWorkers; i++ {
		rawDataChs[i] = make(chan rawEntry, 10)
		resultChs[i] = make(chan *new_substate.Substate, 10)
	}

	// Start i => raw data stage
	i.wg.Add(1)
	go func() {
		defer func() {
			for _, c := range rawDataChs {
				close(c)
			}
			i.wg.Done()
		}()
		step := 0
		for {
			if !i.iter.Next() {
				return
			}

			key := make([]byte, len(i.iter.Key()))
			copy(key, i.iter.Key())
			value := make([]byte, len(i.iter.Value()))
			copy(value, i.iter.Value())

			res := rawEntry{key, value}

			select {
			case e := <-errCh:
				i.err = e
				return
			case rawDataChs[step] <- res: // fall-through
			}
			step = (step + 1) % numWorkers
		}
	}()

	// Start raw data => parsed transaction stage (parallel)
	for w := 0; w < numWorkers; w++ {
		i.wg.Add(1)
		id := w

		go func() {
			defer func() {
				close(resultChs[id])
				i.wg.Done()
			}()
			for raw := range rawDataChs[id] {
				transaction, err := i.decode(raw)
				if err != nil {
					errCh <- err
					return
				}
				resultChs[id] <- transaction
			}
		}()
	}

	// Start the go routine moving transactions from parsers to sink in order
	i.wg.Add(1)
	go func() {
		defer func() {
			close(i.resultCh)
			i.wg.Done()
		}()
		step := 0
		for openProducers := numWorkers; openProducers > 0; {
			next := <-resultChs[step%numWorkers]
			if next != nil {
				i.resultCh <- next
			} else {
				openProducers--
			}
			step++
		}
	}()
}
