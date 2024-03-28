package db

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/rlp"
)

type DestroyedAccountDB struct {
	backend BaseDB
}

func NewDestroyedAccountDB(backend BaseDB) *DestroyedAccountDB {
	return &DestroyedAccountDB{backend: backend}
}

func OpenDestroyedAccountDB(destroyedAccountDir string) (*DestroyedAccountDB, error) {
	return openDestroyedAccountDB(destroyedAccountDir, &opt.Options{ReadOnly: false}, nil, nil)
}

func OpenDestroyedAccountDBReadOnly(destroyedAccountDir string) (*DestroyedAccountDB, error) {
	return openDestroyedAccountDB(destroyedAccountDir, &opt.Options{ReadOnly: true}, nil, nil)
}

func openDestroyedAccountDB(destroyedAccountDir string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*DestroyedAccountDB, error) {
	log.Println("substate: OpenDestroyedAccountDB")
	backend, err := newBaseDB(destroyedAccountDir, o, wo, ro)
	if err != nil {
		return nil, fmt.Errorf("error opening deletion-db %s: %v", destroyedAccountDir, err)
	}
	return NewDestroyedAccountDB(backend), nil
}

func (db *DestroyedAccountDB) Close() error {
	return db.backend.Close()
}

type SuicidedAccountLists struct {
	DestroyedAccounts   []types.Address
	ResurrectedAccounts []types.Address
}

func (db *DestroyedAccountDB) SetDestroyedAccounts(block uint64, tx int, des []types.Address, res []types.Address) error {
	accountList := SuicidedAccountLists{DestroyedAccounts: des, ResurrectedAccounts: res}
	value, err := rlp.EncodeToBytes(accountList)
	if err != nil {
		return err
	}
	return db.backend.Put(encodeDestroyedAccountKey(block, tx), value)
}

func (db *DestroyedAccountDB) GetDestroyedAccounts(block uint64, tx int) ([]types.Address, []types.Address, error) {
	data, err := db.backend.Get(encodeDestroyedAccountKey(block, tx))
	if err != nil {
		return nil, nil, err
	}
	list, err := DecodeAddressList(data)
	return list.DestroyedAccounts, list.ResurrectedAccounts, err
}

// GetAccountsDestroyedInRange get list of all accounts between block from and to (including from and to).
func (db *DestroyedAccountDB) GetAccountsDestroyedInRange(from, to uint64) ([]types.Address, error) {
	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, from)

	iter := db.backend.NewIterator([]byte(DestroyedAccountPrefix), startingBlockBytes)
	defer iter.Release()
	isDestroyed := make(map[types.Address]bool)
	for iter.Next() {
		block, _, err := DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return nil, err
		}
		if block > to {
			break
		}
		list, err := DecodeAddressList(iter.Value())
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

	var accountList []types.Address
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

func DecodeAddressList(data []byte) (SuicidedAccountLists, error) {
	list := SuicidedAccountLists{}
	err := rlp.DecodeBytes(data, &list)
	return list, err
}

// GetFirstKey returns the first block number in the database.
func (db *DestroyedAccountDB) GetFirstKey() (uint64, error) {
	iter := db.backend.NewIterator([]byte(DestroyedAccountPrefix), nil)
	defer iter.Release()

	for iter.Next() {
		firstBlock, _, err := DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}
		return firstBlock, nil
	}
	return 0, fmt.Errorf("no updateset found")
}

// GetLastKey returns the last block number in the database.
// if not found then returns 0
func (db *DestroyedAccountDB) GetLastKey() (uint64, error) {
	var block uint64
	var err error
	iter := db.backend.NewIterator([]byte(DestroyedAccountPrefix), nil)
	for iter.Next() {
		block, _, err = DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %v", err)
		}
	}
	iter.Release()
	return block, nil
}
