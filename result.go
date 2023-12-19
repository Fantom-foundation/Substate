package substate

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Result is the transaction result - hence receipt
type Result struct {
	Status uint64
	Bloom  types.Bloom
	Logs   []*types.Log

	ContractAddress common.Address
	GasUsed         uint64
}

func NewResult(status uint64, bloom types.Bloom, logs []*types.Log, contractAddress common.Address, gasUsed uint64) *Result {
	return &Result{
		Status:          status,
		Bloom:           bloom,
		Logs:            logs,
		ContractAddress: contractAddress,
		GasUsed:         gasUsed,
	}
}
