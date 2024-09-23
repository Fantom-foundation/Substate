package db

import (
	"fmt"

	pb "github.com/Fantom-foundation/Substate/protobuf"
	"github.com/Fantom-foundation/Substate/rlp"
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/golang/protobuf/proto"
)

// SetDecoder sets the runtime parsing behavior of substateDB
// intended usage:
//
//	db := &substateDB{..}           // initializing db
//	        .SetDecoder(<encoding>) // end of init, or right before decoding
func (db *substateDB) SetDecoder(encoding string) *substateDB {
	db.decodeSubstate = getDecoderFunc(encoding, db.GetCode)
	return db
}

type substateDecoder interface {
	DecodeSubstate(bytes []byte, block uint64, tx int) (*substate.Substate, error)
}

// DecodeSubstate implemented by substateDB
func (db *substateDB) DecodeSubstate(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
	if db.decodeSubstate == nil {
		db.SetDecoder("default")
	}
	return db.decodeSubstate(bytes, block, tx)
}

// decoderFunc aliases the common function used to decode substate
type decoderFunc func([]byte, uint64, int) (*substate.Substate, error)

type codeLookup = func(types.Hash) ([]byte, error)

func getDecoderFunc(encoding string, lookup codeLookup) decoderFunc {
	switch encoding {
	case "protobuf", "pb":
		return func(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
			return decodeProtobuf(bytes, lookup, block, tx)
		}
	default:
		fallthrough
	case "rlp":
		return func(bytes []byte, block uint64, tx int) (*substate.Substate, error) {
			return decodeRlp(bytes, lookup, block, tx)
		}
	}
}

// decodeRlp decodes into substate the provided rlp-encoded bytecode
func decodeRlp(bytes []byte, lookup codeLookup, block uint64, tx int) (*substate.Substate, error) {
	rlpSubstate, err := rlp.Decode(bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot decode data into rlp block: %v, tx %v; %w", block, tx, err)
	}

	return rlpSubstate.ToSubstate(lookup, block, tx)
}

// decodeProtobuf decodes into substate the provided protobuf-encoded bytecode
func decodeProtobuf(bytes []byte, lookup codeLookup, block uint64, tx int) (*substate.Substate, error) {
	pbSubstate := &pb.Substate{}
	if err := proto.Unmarshal(bytes, pbSubstate); err != nil {
		return nil, fmt.Errorf("cannot decode data into protobuf block: %v, tx %v; %w", block, tx, err)
	}

	return pbSubstate.Decode(lookup, block, tx)
}
