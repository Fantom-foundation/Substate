package db

import (
	"encoding/binary"
)

const (
	MetadataPrefix       = "md"
	UpdatesetPrefix      = "us"
	UpdatesetIntervalKey = MetadataPrefix + UpdatesetPrefix + "in"
	UpdatesetSizeKey     = MetadataPrefix + UpdatesetPrefix + "si"
)

// PutMetadata into db
func (db *updateDB) PutMetadata(interval, size uint64) error {

	byteInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(byteInterval, interval)

	if err := db.backend.Put([]byte(UpdatesetIntervalKey), byteInterval, db.wo); err != nil {
		return err
	}

	sizeInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeInterval, size)

	if err := db.backend.Put([]byte(UpdatesetSizeKey), sizeInterval, db.wo); err != nil {
		return err
	}

	return nil
}

// GetMetadata from db
func (db *updateDB) GetMetadata() (uint64, uint64, error) {
	byteInterval, err := db.backend.Get([]byte(UpdatesetIntervalKey), db.ro)
	if err != nil {
		return 0, 0, err
	}

	byteSize, err := db.backend.Get([]byte(UpdatesetSizeKey), db.ro)
	if err != nil {
		return 0, 0, err
	}

	return binary.BigEndian.Uint64(byteInterval), binary.BigEndian.Uint64(byteSize), nil
}
