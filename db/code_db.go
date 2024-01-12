package db

import (
	"errors"
	"fmt"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/crypto"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const Stage1CodePrefix = "1c" // Stage1CodePrefix + codeHash (256-bit) -> code

// CodeDB is a wrappe around BaseDB. It extends it with Has/Get/PutCode functions.
type CodeDB interface {
	BaseDB

	// HasCode returns true if the DB does contain given code hash.
	HasCode(common.Hash) (bool, error)

	// GetCode gets the code for the given hash.
	GetCode(common.Hash) ([]byte, error)

	// PutCode creates hash for given code and inserts it into the DB.
	PutCode([]byte) error

	// DeleteCode deletes the code for given hash.
	DeleteCode(common.Hash) error
}

// NewDefaultCodeDB creates new instance of CodeDB with default options.
func NewDefaultCodeDB(path string) (CodeDB, error) {
	return newCodeDB(path, nil, nil, nil)
}

// NewCodeDB creates new instance of CodeDB with customizable options.
// Note: Any of three options is nillable. If that's the case a default value for the option is set.
func NewCodeDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (CodeDB, error) {
	return newCodeDB(path, o, wo, ro)
}

func newCodeDB(path string, o *opt.Options, wo *opt.WriteOptions, ro *opt.ReadOptions) (*codeDB, error) {
	base, err := newBaseDB(path, o, wo, ro)
	if err != nil {
		return nil, err
	}
	return &codeDB{base}, nil
}

type codeDB struct {
	*baseDB
}

var ErrorEmptyHash = errors.New("give hash is empty")

// HasCode returns true if the baseDB does contain given code hash.
func (db *codeDB) HasCode(codeHash common.Hash) (bool, error) {
	emptyHash := common.Hash{}
	if codeHash == emptyHash {
		return false, ErrorEmptyHash
	}

	key := Stage1CodeKey(codeHash)
	has, err := db.Has(key)
	if err != nil {
		return false, err
	}
	return has, nil
}

// GetCode gets the code for the given hash.
func (db *codeDB) GetCode(codeHash common.Hash) ([]byte, error) {
	if codeHash == common.EmptyHash {
		return nil, ErrorEmptyHash
	}

	key := Stage1CodeKey(codeHash)
	code, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get code %s: %v", codeHash.Hex(), err)
	}
	return code, nil
}

// PutCode creates hash for given code and inserts it into the baseDB.
func (db *codeDB) PutCode(code []byte) error {
	codeHash := crypto.Keccak256Hash(code)
	key := Stage1CodeKey(codeHash)
	err := db.Put(key, code)
	if err != nil {
		return fmt.Errorf("cannot put code %v; %v", codeHash.Hex(), err)
	}

	return nil
}

// DeleteCode deletes the code for the given hash.
func (db *codeDB) DeleteCode(codeHash common.Hash) error {
	if codeHash == common.EmptyHash {
		return ErrorEmptyHash
	}

	key := Stage1CodeKey(codeHash)
	err := db.Delete(key)
	if err != nil {
		return fmt.Errorf("cannot get code %s: %v", codeHash.Hex(), err)
	}
	return nil
}

// Stage1CodeKey returns Stage1CodePrefix with appended
// codeHash creating key used in baseDB for Codes.
func Stage1CodeKey(codeHash common.Hash) []byte {
	prefix := []byte(Stage1CodePrefix)
	return append(prefix, codeHash.Bytes()...)
}

// DecodeStage1CodeKey decodes key created by Stage1CodeKey back to hash.
func DecodeStage1CodeKey(key []byte) (codeHash common.Hash, err error) {
	prefix := Stage1CodePrefix
	if len(key) != len(prefix)+32 {
		err = fmt.Errorf("invalid length of stage1 code key: %v", len(key))
		return
	}
	if p := string(key[:2]); p != prefix {
		err = fmt.Errorf("invalid prefix of stage1 code key: %#x", p)
		return
	}
	codeHash = common.BytesToHash(key[len(prefix):])
	return
}
