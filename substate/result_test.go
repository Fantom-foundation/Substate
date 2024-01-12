package substate

import (
	"testing"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/types"
)

func TestAccount_EqualStatus(t *testing.T) {
	res := &Result{Status: 0}
	comparedRes := &Result{Status: 1}

	if res.Equal(comparedRes) {
		t.Fatal("results status are different but equal returned true")
	}

	comparedRes.Status = res.Status
	if !res.Equal(comparedRes) {
		t.Fatal("results status are same but equal returned false")
	}
}

func TestAccount_EqualBloom(t *testing.T) {
	res := &Result{Bloom: types.Bloom{0}}
	comparedRes := &Result{Bloom: types.Bloom{1}}

	if res.Equal(comparedRes) {
		t.Fatal("results Bloom are different but equal returned true")
	}

	comparedRes.Bloom = res.Bloom
	if !res.Equal(comparedRes) {
		t.Fatal("results Bloom are same but equal returned false")
	}
}

func TestAccount_EqualLogs(t *testing.T) {
	res := &Result{Logs: []*types.Log{{Address: common.Address{0}}}}
	comparedRes := &Result{Logs: []*types.Log{{Address: common.Address{1}}}}

	if res.Equal(comparedRes) {
		t.Fatal("results Log are different but equal returned true")
	}

	comparedRes.Logs = res.Logs
	if !res.Equal(comparedRes) {
		t.Fatal("results Log are same but equal returned false")
	}
}

func TestAccount_EqualContractAddress(t *testing.T) {
	res := &Result{ContractAddress: common.Address{0}}
	comparedRes := &Result{ContractAddress: common.Address{1}}

	if res.Equal(comparedRes) {
		t.Fatal("results ContractAddress are different but equal returned true")
	}

	comparedRes.ContractAddress = res.ContractAddress
	if !res.Equal(comparedRes) {
		t.Fatal("results ContractAddress are same but equal returned false")
	}
}

func TestAccount_EqualGasUsed(t *testing.T) {
	res := &Result{GasUsed: 0}
	comparedRes := &Result{GasUsed: 1}

	if res.Equal(comparedRes) {
		t.Fatal("results GasUsed are different but equal returned true")
	}

	comparedRes.GasUsed = res.GasUsed
	if !res.Equal(comparedRes) {
		t.Fatal("results GasUsed are same but equal returned false")
	}
}
