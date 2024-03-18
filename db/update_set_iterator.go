package db

import (
	"encoding/binary"
	"fmt"

	"github.com/Fantom-foundation/Substate/types/rlp"
	"github.com/Fantom-foundation/Substate/updateset"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func newUpdateSetIterator(db *updateDB, start, end uint64) *updateSetIterator {
	num := make([]byte, 4)
	binary.BigEndian.PutUint32(num, uint32(start))

	r := util.BytesPrefix([]byte(UpdateDBPrefix))
	r.Start = append(r.Start, num...)

	return &updateSetIterator{
		iterator: newIterator[*updateset.UpdateSet](db.backend.NewIterator(r, db.ro)),
		db:       db,
		endBlock: end,
	}
}

type updateSetIterator struct {
	iterator[*updateset.UpdateSet]
	db       *updateDB
	endBlock uint64
}

func (i *updateSetIterator) decode(data rawEntry) (*updateset.UpdateSet, error) {
	key := data.key
	value := data.value

	block, err := DecodeUpdateSetKey(data.key)
	if err != nil {
		panic(fmt.Errorf("substate: invalid update-set key found: %v - issue: %v", key, err))
	}

	var updateSetRLP updateset.UpdateSetRLP
	err = rlp.DecodeBytes(value, &updateSetRLP)
	if err != nil {
		return nil, err
	}

	updateSet, err := updateSetRLP.ToWorldState(i.db.GetCode, block)
	if err != nil {
		return nil, err

	}

	return &updateset.UpdateSet{
		Block:           block,
		WorldState:      updateSet.WorldState,
		DeletedAccounts: updateSetRLP.DeletedAccounts,
	}, nil
}

func (i *updateSetIterator) start(_ int) {
	i.wg.Add(1)

	go func() {
		defer func() {
			close(i.resultCh)
			i.wg.Done()
		}()
		for {
			if !i.iter.Next() {
				return
			}

			key := make([]byte, len(i.iter.Key()))
			copy(key, i.iter.Key())

			// Decode key, if past the end block, stop here.
			// This avoids filling channels which huge data objects that are not consumed.
			block, err := DecodeUpdateSetKey(key)
			if err != nil {
				i.err = err
				return
			}
			if block > i.endBlock {
				return
			}

			value := make([]byte, len(i.iter.Value()))
			copy(value, i.iter.Value())

			raw := rawEntry{key, value}

			us, err := i.decode(raw)
			if err != nil {
				i.err = err
				return
			}

			i.resultCh <- us
		}
	}()
}
