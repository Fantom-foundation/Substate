package substate

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/types"
)

func TestEnv_EqualCoinbase(t *testing.T) {
	env := &Env{
		Coinbase: types.Address{0},
	}
	comparedEnv := &Env{
		Coinbase: types.Address{1},
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs coinbase are different but equal returned true")
	}

	comparedEnv.Coinbase = env.Coinbase
	if !env.Equal(comparedEnv) {
		t.Fatal("envs coinbase are same but equal returned false")
	}
}

func TestEnv_EqualDifficulty(t *testing.T) {
	env := &Env{
		Difficulty: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		Difficulty: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs difficulty are different but equal returned true")
	}

	comparedEnv.Difficulty = env.Difficulty
	if !env.Equal(comparedEnv) {
		t.Fatal("envs difficulty are same but equal returned false")
	}
}

func TestEnv_EqualGasLimit(t *testing.T) {
	env := &Env{
		GasLimit: 0,
	}
	comparedEnv := &Env{
		GasLimit: 1,
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs gasLimit are different but equal returned true")
	}

	comparedEnv.GasLimit = env.GasLimit
	if !env.Equal(comparedEnv) {
		t.Fatal("envs gasLimit are same but equal returned false")
	}
}

func TestEnv_EqualNumber(t *testing.T) {
	env := &Env{
		Number: 0,
	}
	comparedEnv := &Env{
		Number: 1,
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs number are different but equal returned true")
	}

	comparedEnv.Number = env.Number
	if !env.Equal(comparedEnv) {
		t.Fatal("envs number are same but equal returned false")
	}
}

func TestEnv_EqualBlockHashes(t *testing.T) {
	env := &Env{
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{0})},
	}
	comparedEnv := &Env{
		BlockHashes: map[uint64]types.Hash{0: types.BytesToHash([]byte{1})},
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs hashes for block 0 are different but equal returned true")
	}

	comparedEnv.BlockHashes = map[uint64]types.Hash{1: types.BytesToHash([]byte{1})}

	if env.Equal(comparedEnv) {
		t.Fatal("envs blockHashes are different but equal returned true")
	}

	comparedEnv.BlockHashes = env.BlockHashes
	if !env.Equal(comparedEnv) {
		t.Fatal("envs number are same but equal returned false")
	}
}

func TestEnv_EqualBaseFee(t *testing.T) {
	env := &Env{
		BaseFee: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		BaseFee: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs BaseFee are different but equal returned true")
	}

	comparedEnv.BaseFee = env.BaseFee
	if !env.Equal(comparedEnv) {
		t.Fatal("envs BaseFee are same but equal returned false")
	}
}

func TestEnv_EqualBlobBaseFee(t *testing.T) {
	env := &Env{
		BlobBaseFee: new(big.Int).SetUint64(0),
	}
	comparedEnv := &Env{
		BlobBaseFee: new(big.Int).SetUint64(1),
	}

	if env.Equal(comparedEnv) {
		t.Fatal("envs BlobBaseFee are different but equal returned true")
	}

	comparedEnv.BlobBaseFee = env.BlobBaseFee
	if !env.Equal(comparedEnv) {
		t.Fatal("envs BlobBaseFee are same but equal returned false")
	}
}
