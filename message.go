package substate

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// AccessList determines to which storage keys (values) has address (key) access
type AccessList map[common.Address][]common.Hash

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
