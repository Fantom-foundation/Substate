package rlp

import (
	"reflect"
)

// byteArrayBytes returns a slice of the byte array v.
func byteArrayBytes(v reflect.Value) []byte {
	return v.Slice(0, v.Len()).Bytes()
}
