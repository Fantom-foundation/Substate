package protobuf

import (
	"math/big"

	"github.com/Fantom-foundation/Substate/types"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

func BytesValueToHash(bv *wrapperspb.BytesValue) *types.Hash {
	if bv == nil {
		return nil
	}
	hash := types.BytesToHash(bv.GetValue())
	return &hash
}

func BytesValueToBigInt(bv *wrapperspb.BytesValue) *big.Int {
	if bv == nil {
		return nil
	}
	return new(big.Int).SetBytes(bv.GetValue())
}

func BytesValueToAddress(bv *wrapperspb.BytesValue) *types.Address {
	if bv == nil {
		return nil
	}
	addr := types.BytesToAddress(bv.GetValue())
	return &addr
}
