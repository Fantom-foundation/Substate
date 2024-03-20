package types

// AccessList is an EIP-2930 access list.
type AccessList []AccessTuple

// AccessTuple is the element type of an access list.
type AccessTuple struct {
	Address     Address `json:"address"        gencodec:"required"`
	StorageKeys []Hash  `json:"storageKeys"    gencodec:"required"`
}
