package types

import (
	"encoding/hex"
)

// Address represents the 20 byte address of an Ethereum account.
type Address [20]byte

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}
