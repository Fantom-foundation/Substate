package substate

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/rlp"
)

type DestroyedAccountDB struct {
	backend BackendDatabase
}

func NewDestroyedAccountDB(backend BackendDatabase) *DestroyedAccountDB {
	return &DestroyedAccountDB{backend: backend}
}

func OpenDestroyedAccountDB(destroyedAccountDir string) (*DestroyedAccountDB, error) {
	return openDestroyedAccountDB(destroyedAccountDir, false)
}

func OpenDestroyedAccountDBReadOnly(destroyedAccountDir string) (*DestroyedAccountDB, error) {
	return openDestroyedAccountDB(destroyedAccountDir, true)
}

func openDestroyedAccountDB(destroyedAccountDir string, readOnly bool) (*DestroyedAccountDB, error) {
	log.Println("substate: OpenDestroyedAccountDB")
	backend, err := rawdb.NewLevelDBDatabase(destroyedAccountDir, 1024, 100, "destroyed_accounts", readOnly)
	if err != nil {
		return nil, fmt.Errorf("error opening deletion-db %s: %v", destroyedAccountDir, err)
	}
	return NewDestroyedAccountDB(backend), nil
}

func (db *DestroyedAccountDB) Close() error {
	return db.backend.Close()
}

type SuicidedAccountLists struct {
	DestroyedAccounts   []common.Address
	ResurrectedAccounts []common.Address
}

func (db *DestroyedAccountDB) SetDestroyedAccounts(block uint64, tx int, des []common.Address, res []common.Address) error {
	accountList := SuicidedAccountLists{DestroyedAccounts: des, ResurrectedAccounts: res}
	value, err := rlp.EncodeToBytes(accountList)
	if err != nil {
		panic(err)
	}
	return db.backend.Put(encodeDestroyedAccountKey(block, tx), value)
}

func (db *DestroyedAccountDB) GetDestroyedAccounts(block uint64, tx int) ([]common.Address, []common.Address, error) {
	data, err := db.backend.Get(encodeDestroyedAccountKey(block, tx))
	if err != nil {
		return nil, nil, err
	}
	list, err := decodeAddressList(data)
	return list.DestroyedAccounts, list.ResurrectedAccounts, err
}

func (db *DestroyedAccountDB) getFirstKeyInBlock(block uint64) []byte {
	prefix := []byte(DestroyedAccountPrefix)
	blockBytes := make([]byte, len(prefix)+8)
	copy(blockBytes[0:], prefix)
	binary.BigEndian.PutUint64(blockBytes[len(prefix):], block)
	iter := db.backend.NewIterator(blockBytes, nil)
	defer iter.Release()
	if iter.Key() != nil {
		return iter.Key()[len(prefix):]
	}
	return nil
}

// GetAccountsDestroyedInRange get list of all accounts between block from and to (including from and to).
func (db *DestroyedAccountDB) GetAccountsDestroyedInRange(from, to uint64) ([]common.Address, error) {
	firstKey := db.getFirstKeyInBlock(from)
	iter := db.backend.NewIterator([]byte(DestroyedAccountPrefix), firstKey)
	defer iter.Release()
	isDestroyed := make(map[common.Address]bool)
	for iter.Next() {
		block, _, err := DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return nil, err
		}
		if block > to {
			break
		}
		list, err := decodeAddressList(iter.Value())
		if err != nil {
			return nil, err
		}
		for _, addr := range list.DestroyedAccounts {
			isDestroyed[addr] = true
		}
		for _, addr := range list.ResurrectedAccounts {
			isDestroyed[addr] = false
		}
	}

	accountList := []common.Address{}
	for addr, isDeleted := range isDestroyed {
		if isDeleted {
			accountList = append(accountList, addr)
		}
	}
	return accountList, nil
}

const (
	DestroyedAccountPrefix = "da" // DestroyedAccountPrefix + block (64-bit) -> SuicidedAccountLists
)

func encodeDestroyedAccountKey(block uint64, tx int) []byte {
	prefix := []byte(DestroyedAccountPrefix)
	key := make([]byte, len(prefix)+12)
	copy(key[0:], prefix)
	binary.BigEndian.PutUint64(key[len(prefix):], block)
	binary.BigEndian.PutUint32(key[len(prefix)+8:], uint32(tx))
	return key
}

func DecodeDestroyedAccountKey(data []byte) (uint64, int, error) {
	if len(data) != len(DestroyedAccountPrefix)+12 {
		return 0, 0, fmt.Errorf("invalid length of destroyed account key, expected %d, got %d", len(DestroyedAccountPrefix)+12, len(data))
	}
	if string(data[0:len(DestroyedAccountPrefix)]) != DestroyedAccountPrefix {
		return 0, 0, fmt.Errorf("invalid prefix of destroyed account key")
	}
	block := binary.BigEndian.Uint64(data[len(DestroyedAccountPrefix):])
	tx := binary.BigEndian.Uint32(data[len(DestroyedAccountPrefix)+8:])
	return block, int(tx), nil
}

func decodeAddressList(data []byte) (SuicidedAccountLists, error) {
	list := SuicidedAccountLists{}
	err := rlp.DecodeBytes(data, &list)
	return list, err
}
