package substate

import (
	"bytes"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/hash"
)

type Message struct {
	Nonce      uint64
	CheckNonce bool // inversion of IsFake
	GasPrice   *big.Int
	Gas        uint64

	From  types.Address
	To    *types.Address // nil means contract creation
	Value *big.Int
	Data  []byte

	// for memoization
	dataHash *types.Hash

	// Berlin hard fork, EIP-2930: Optional access lists
	AccessList types.AccessList // nil if EIP-2930 is not activated

	// London hard fork, EIP-1559: Fee market
	GasFeeCap *big.Int // GasPrice if EIP-1559 is not activated
	GasTipCap *big.Int // GasPrice if EIP-1559 is not activated

	// Cancun hard fork, EIP-4844
	BlobGasFeeCap *big.Int
	BlobHashes    []types.Hash
}

func NewMessage(
	nonce uint64,
	checkNonce bool,
	gasPrice *big.Int,
	gas uint64,
	from types.Address,
	to *types.Address,
	value *big.Int,
	data []byte,
	dataHash *types.Hash,
	accessList types.AccessList,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	blobGasFeeCap *big.Int,
	blobHashes []types.Hash,
) *Message {
	return &Message{
		Nonce:         nonce,
		CheckNonce:    checkNonce,
		GasPrice:      gasPrice,
		Gas:           gas,
		From:          from,
		To:            to,
		Value:         value,
		Data:          data,
		dataHash:      dataHash,
		AccessList:    accessList,
		GasFeeCap:     gasFeeCap,
		GasTipCap:     gasTipCap,
		BlobGasFeeCap: blobGasFeeCap,
		BlobHashes:    blobHashes,
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
		m.GasTipCap.Cmp(y.GasTipCap) == 0 &&
		m.BlobGasFeeCap.Cmp(y.BlobGasFeeCap) == 0
	if !equal {
		return false
	}

	if !slices.Equal(m.BlobHashes, y.BlobHashes) {
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
func (m *Message) DataHash() types.Hash {
	if m.dataHash == nil {
		dataHash := hash.Keccak256Hash(m.Data)
		m.dataHash = &dataHash
	}
	return *m.dataHash
}

func (m *Message) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Nonce: %v\n", m.Nonce))
	builder.WriteString(fmt.Sprintf("CheckNonce: %v\n", m.CheckNonce))
	builder.WriteString(fmt.Sprintf("From: %s\n", m.From))
	builder.WriteString(fmt.Sprintf("To: %s\n", m.To))
	builder.WriteString(fmt.Sprintf("Value: %v\n", m.Value.String()))
	builder.WriteString(fmt.Sprintf("Data: %v\n", string(m.Data)))
	builder.WriteString(fmt.Sprintf("Data Hash: %s\n", m.dataHash))
	builder.WriteString(fmt.Sprintf("Gas Fee Cap: %v\n", m.GasFeeCap.String()))
	builder.WriteString(fmt.Sprintf("Gas Tip Cap: %v\n", m.GasTipCap.String()))

	for _, tuple := range m.AccessList {
		builder.WriteString(fmt.Sprintf("Address: %s", tuple.Address))
		for i, key := range tuple.StorageKeys {
			builder.WriteString(fmt.Sprintf("Storage Key %v: %v", i, key))
		}
	}

	return builder.String()
}
