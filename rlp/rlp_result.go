package rlp

import (
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/common"
)

func NewResult(result *substate.Result) *Result {
	return &Result{
		Status:          result.Status,
		Bloom:           result.Bloom,
		Logs:            result.Logs,
		ContractAddress: result.ContractAddress,
		GasUsed:         result.GasUsed,
	}
}

type Result struct {
	Status uint64
	Bloom  [256]byte
	Logs   []*types.Log

	ContractAddress common.Address
	GasUsed         uint64
}

// ToSubstate transforms r from Result to substate.Result.
func (r Result) ToSubstate() *substate.Result {
	return &substate.Result{
		Status:          r.Status,
		Bloom:           r.Bloom,
		Logs:            r.Logs,
		ContractAddress: r.ContractAddress,
		GasUsed:         r.GasUsed,
	}
}
