package db

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/hash"
)

const CodeDBPrefix = "1c" // CodeDBPrefix + codeHash (256-bit) -> code

// CodeDB is a wrappe around BaseDB. It extends it with Has/Get/PutCode functions.
type CodeDB interface {
	BaseDB

	// HasCode returns true if the DB does contain given code hash.
	HasCode(types.Hash) (bool, error)

	// GetCode gets the code for the given hash.
	GetCode(types.Hash) ([]byte, error)

	// PutCode creates hash for given code and inserts it into the DB.
	PutCode([]byte) error

	// DeleteCode deletes the code for given hash.
	DeleteCode(types.Hash) error
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

func MakeDefaultCodeDBFromBaseDB(db BaseDB) CodeDB {
	return &codeDB{&baseDB{backend: db.getBackend()}}
}

// NewReadOnlyCodeDB creates a new instance of read-only CodeDB.
func NewReadOnlyCodeDB(path string) (CodeDB, error) {
	return newCodeDB(path, &opt.Options{ReadOnly: true}, nil, nil)
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
func (db *codeDB) HasCode(codeHash types.Hash) (bool, error) {
	if codeHash.IsEmpty() {
		return false, ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	has, err := db.Has(key)
	if err != nil {
		return false, err
	}
	return has, nil
}

// GetCode gets the code for the given hash.
func (db *codeDB) GetCode(codeHash types.Hash) ([]byte, error) {
	if codeHash.IsEmpty() {
		return nil, ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	code, err := db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get code %s: %w", codeHash, err)
	}

	return code, nil
}

// PutCode creates hash for given code and inserts it into the baseDB.
func (db *codeDB) PutCode(code []byte) error {
	codeHash := hash.Keccak256Hash(code)
	key := CodeDBKey(codeHash)
	err := db.Put(key, code)
	if err != nil {
		return fmt.Errorf("cannot put code %s: %w", codeHash, err)
	}

	return nil
}

// DeleteCode deletes the code for the given hash.
func (db *codeDB) DeleteCode(codeHash types.Hash) error {
	if codeHash.IsEmpty() {
		return ErrorEmptyHash
	}

	key := CodeDBKey(codeHash)
	err := db.Delete(key)
	if err != nil {
		return fmt.Errorf("cannot delete code %s: %w", codeHash, err)
	}
	return nil
}

// CodeDBKey returns CodeDBPrefix with appended
// codeHash creating key used in baseDB for Codes.
func CodeDBKey(codeHash types.Hash) []byte {
	prefix := []byte(CodeDBPrefix)
	return append(prefix, codeHash[:]...)
}

// DecodeCodeDBKey decodes key created by CodeDBKey back to hash.
func DecodeCodeDBKey(key []byte) (codeHash types.Hash, err error) {
	prefix := CodeDBPrefix
	if len(key) != len(prefix)+32 {
		err = fmt.Errorf("invalid length of code db key: %v", len(key))
		return
	}
	if p := string(key[:2]); p != prefix {
		err = fmt.Errorf("invalid prefix of code db key: %#x", p)
		return
	}
	var h types.Hash
	h.SetBytes(key[len(prefix):])
	codeHash = types.BytesToHash(key[len(prefix):])
	return
}
