package substate

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// AccessList is an EIP-2930 access list.
type AccessList []AccessTuple

// AccessTuple is the element type of access list.
type AccessTuple struct {
	Address     common.Address `json:"address"        gencodec:"required"`
	StorageKeys []common.Hash  `json:"storageKeys"    gencodec:"required"`
}

type Message struct {
	Nonce      uint64
	CheckNonce bool // inversion of IsFake
	GasPrice   *big.Int
	Gas        uint64

	From  common.Address
	To    *common.Address // nil means contract creation
	Value *big.Int
	Data  []byte

	// for memoization
	dataHash *common.Hash

	// Berlin hard fork, EIP-2930: Optional access lists
	AccessList AccessList // nil if EIP-2930 is not activated

	// London hard fork, EIP-1559: Fee market
	GasFeeCap *big.Int // GasPrice if EIP-1559 is not activated
	GasTipCap *big.Int // GasPrice if EIP-1559 is not activated
}

func NewMessage(nonce uint64, checkNonce bool, gasPrice *big.Int, gas uint64, from common.Address, to *common.Address, value *big.Int, data []byte, dataHash *common.Hash, accessList AccessList, gasFeeCap *big.Int, gasTipCap *big.Int) *Message {
	return &Message{
		Nonce:      nonce,
		CheckNonce: checkNonce,
		GasPrice:   gasPrice,
		Gas:        gas, From: from,
		To: to, Value: value,
		Data:       data,
		dataHash:   dataHash,
		AccessList: accessList,
		GasFeeCap:  gasFeeCap,
		GasTipCap:  gasTipCap,
	}
}

// Equal returns true if m is y or if values of m are equal to values of y.
// Otherwise, m and y are not equal hence false is returned.
func (m *Message) Equal(y *Message) bool {
	if m == y {
		return true
	}

	if (m == nil || y == nil) && m != y {
		return false
	}

	// check values
	equal := m.Nonce == y.Nonce &&
		m.CheckNonce == y.CheckNonce &&
		m.GasPrice.Cmp(y.GasPrice) == 0 &&
		m.Gas == y.Gas &&
		m.From == y.From &&
		(m.To == y.To || (m.To != nil && y.To != nil && *m.To == *y.To)) &&
		m.Value.Cmp(y.Value) == 0 &&
		bytes.Equal(m.Data, y.Data) &&
		len(m.AccessList) == len(y.AccessList) &&
		m.GasFeeCap.Cmp(y.GasFeeCap) == 0 &&
		m.GasTipCap.Cmp(y.GasTipCap) == 0
	if !equal {
		return false
	}

	// check AccessList
	for i, mTuple := range m.AccessList {
		yTuple := y.AccessList[i]

		// check addresses position
		if mTuple.Address != yTuple.Address {
			return false
		}

		// check size of StorageKeys
		if len(mTuple.StorageKeys) != len(yTuple.StorageKeys) {
			return false
		}

		// check StorageKeys
		for j, mKey := range mTuple.StorageKeys {
			yKey := yTuple.StorageKeys[j]
			if mKey != yKey {
				return false
			}
		}
	}

	return true
}

// DataHash returns m.dataHash if it exists. If not, it is generated using Keccak256 algorithm.
func (m *Message) DataHash() common.Hash {
	if m.dataHash == nil {
		dataHash := crypto.Keccak256Hash(m.Data)
		m.dataHash = &dataHash
	}
	return *m.dataHash
}
