package rlp

import "reflect"

// RawValue represents an encoded RLP value and can be used to delay
// RLP decoding or to precompute an encoding. Note that the decoder does
// not verify whether the content of RawValues is valid RLP.
type RawValue []byte

var rawValueType = reflect.TypeOf(RawValue{})
