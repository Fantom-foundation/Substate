package substate

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

const MetadataPrefix = "md" // used for interval with size of UpdateSet

type MetadataType byte

const (
	UpdatesetInterval MetadataType = iota
	UpdatesetSize
)

// metadataKey return db key for given MetadataType
func metadataKey(typ MetadataType) ([]byte, error) {
	var (
		byteType []byte
		err      error
	)

	switch typ {
	case UpdatesetInterval:
		byteType, err = rlp.EncodeToBytes("interval")
	case UpdatesetSize:
		byteType, err = rlp.EncodeToBytes("size")
	default:
		err = fmt.Errorf("unknown metadata type")
	}

	if err != nil {
		return nil, err
	}

	return append([]byte(MetadataPrefix), byteType...), nil
}

// PutMetadata into db
func (db *UpdateDB) PutMetadata(typ MetadataType, val uint64) error {

	key, err := metadataKey(typ)
	if err != nil {
		return err
	}

	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, val)

	return db.backend.Put(key, value)
}

// GetMetadata from db
func (db *UpdateDB) GetMetadata() (uint64, uint64, error) {
	// retrieve interval
	intervalKey, err := metadataKey(UpdatesetInterval)
	if err != nil {
		return 0, 0, err
	}
	byteInterval, err := db.backend.Get(intervalKey)
	if err != nil {
		return 0, 0, err
	}

	// retrieve size
	sizeKey, err := metadataKey(UpdatesetSize)
	if err != nil {
		return 0, 0, err
	}
	byteSize, err := db.backend.Get(sizeKey)

	return binary.BigEndian.Uint64(byteInterval), binary.BigEndian.Uint64(byteSize), nil
}
