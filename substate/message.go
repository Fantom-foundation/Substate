package substate

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/crypto"
	"github.com/Fantom-foundation/Substate/geth/types"
)

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
	AccessList types.AccessList // nil if EIP-2930 is not activated

	// London hard fork, EIP-1559: Fee market
	GasFeeCap *big.Int // GasPrice if EIP-1559 is not activated
	GasTipCap *big.Int // GasPrice if EIP-1559 is not activated
}

func NewMessage(
	nonce uint64,
	checkNonce bool,
	gasPrice *big.Int,
	gas uint64,
	from common.Address,
	to *common.Address,
	value *big.Int,
	data []byte,
	dataHash *common.Hash,
	accessList types.AccessList,
	gasFeeCap *big.Int,
	gasTipCap *big.Int) *Message {
	return &Message{
		Nonce:      nonce,
		CheckNonce: checkNonce,
		GasPrice:   gasPrice,
		Gas:        gas,
		From:       from,
		To:         to,
		Value:      value,
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

func (m *Message) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Nonce: %v", m.Nonce))
	builder.WriteString(fmt.Sprintf("CheckNonce: %v", m.CheckNonce))
	builder.WriteString(fmt.Sprintf("From: %v", m.From.Hex()))
	builder.WriteString(fmt.Sprintf("To: %v", m.To.Hex()))
	builder.WriteString(fmt.Sprintf("Value: %v", m.Value.String()))
	builder.WriteString(fmt.Sprintf("Data: %v", string(m.Data)))
	builder.WriteString(fmt.Sprintf("Data Hash: %v", m.dataHash.Hex()))
	builder.WriteString(fmt.Sprintf("Gas Fee Cap: %v", m.GasFeeCap.String()))
	builder.WriteString(fmt.Sprintf("Gas Tip Cap: %v", m.GasTipCap.String()))

	for _, tuple := range m.AccessList {
		builder.WriteString(fmt.Sprintf("Address: %v", tuple.Address.Hex()))
		for i, key := range tuple.StorageKeys {
			builder.WriteString(fmt.Sprintf("Storage Key %v: %v", i, key.Hex()))
		}
	}

	return builder.String()
}
