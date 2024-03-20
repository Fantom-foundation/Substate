package substate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/common"
)

// Result is the transaction result - hence receipt
type Result struct {
	Status          uint64
	Bloom           [256]byte
	Logs            []*types.Log
	ContractAddress common.Address
	GasUsed         uint64
}

func NewResult(status uint64, bloom [256]byte, logs []*types.Log, contractAddress common.Address, gasUsed uint64) *Result {
	return &Result{
		Status:          status,
		Bloom:           bloom,
		Logs:            logs,
		ContractAddress: contractAddress,
		GasUsed:         gasUsed,
	}
}

// Equal returns true if r is y or if values of r are equal to values of y.
// Otherwise, r and y are not equal hence false is returned.
func (r *Result) Equal(y *Result) bool {
	if r == y {
		return true
	}

	if (r == nil || y == nil) && r != y {
		return false
	}

	equal := r.Status == y.Status &&
		r.Bloom == y.Bloom &&
		len(r.Logs) == len(y.Logs) &&
		r.ContractAddress == y.ContractAddress &&
		r.GasUsed == y.GasUsed
	if !equal {
		return false
	}

	for i, logs := range r.Logs {
		yLogs := y.Logs[i]

		equal := logs.Address == yLogs.Address &&
			len(logs.Topics) == len(yLogs.Topics) &&
			bytes.Equal(logs.Data, yLogs.Data)
		if !equal {
			return false
		}

		for i, xt := range logs.Topics {
			yt := yLogs.Topics[i]
			if xt != yt {
				return false
			}
		}
	}

	return true
}

func (r *Result) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Status: %v", r.Status))
	builder.WriteString(fmt.Sprintf("Bloom: %s", r.Bloom))
	builder.WriteString(fmt.Sprintf("Contract Address: %s", r.ContractAddress))
	builder.WriteString(fmt.Sprintf("Gas Used: %v", r.GasUsed))

	for _, log := range r.Logs {
		builder.WriteString(fmt.Sprintf("%v", *log))
	}

	return builder.String()
}
