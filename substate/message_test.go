package substate

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/geth/common"
	"github.com/Fantom-foundation/Substate/geth/crypto"
	"github.com/Fantom-foundation/Substate/geth/types"
)

func TestMessage_EqualNonce(t *testing.T) {
	msg := &Message{Nonce: 0}
	comparedMsg := &Message{Nonce: 1}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages nonce are different but equal returned true")
	}

	comparedMsg.Nonce = msg.Nonce
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages nonce are same but equal returned false")
	}
}

func TestMessage_EqualCheckNonce(t *testing.T) {
	msg := &Message{CheckNonce: false}
	comparedMsg := &Message{CheckNonce: true}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages CheckNonce are different but equal returned true")
	}

	comparedMsg.CheckNonce = msg.CheckNonce
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages CheckNonce are same but equal returned false")
	}
}

func TestMessage_EqualGasPrice(t *testing.T) {
	msg := &Message{GasPrice: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasPrice: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasPrice are different but equal returned true")
	}

	comparedMsg.GasPrice = msg.GasPrice
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasPrice are same but equal returned false")
	}
}

func TestMessage_EqualFrom(t *testing.T) {
	msg := &Message{From: common.Address{0}}
	comparedMsg := &Message{From: common.Address{1}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages From are different but equal returned true")
	}

	comparedMsg.From = msg.From
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages From are same but equal returned false")
	}
}

func TestMessage_EqualTo(t *testing.T) {
	msg := &Message{To: &common.Address{0}}
	comparedMsg := &Message{To: &common.Address{1}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages To are different but equal returned true")
	}

	comparedMsg.To = msg.To
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages To are same but equal returned false")
	}
}

func TestMessage_EqualValue(t *testing.T) {
	msg := &Message{Value: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{Value: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages values are different but equal returned true")
	}

	comparedMsg.Value = msg.Value
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages Value are same but equal returned false")
	}
}

func TestMessage_Equal_DataHashDoesNotAffectResult(t *testing.T) {
	msg := &Message{dataHash: new(common.Hash)}
	*msg.dataHash = common.BytesToHash([]byte{0})
	comparedMsg := &Message{dataHash: new(common.Hash)}
	*comparedMsg.dataHash = common.BytesToHash([]byte{1})

	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal even if it is different")
	}

	comparedMsg.dataHash = msg.dataHash
	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal")
	}
}

func TestMessage_EqualAccessList(t *testing.T) {
	msg := &Message{AccessList: []types.AccessTuple{{Address: common.Address{0}, StorageKeys: []common.Hash{common.BytesToHash([]byte{0})}}}}
	comparedMsg := &Message{AccessList: []types.AccessTuple{{Address: common.Address{0}, StorageKeys: []common.Hash{common.BytesToHash([]byte{1})}}}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Key for same address but equal returned true")
	}

	comparedMsg.AccessList = append(comparedMsg.AccessList, types.AccessTuple{Address: common.Address{0}, StorageKeys: []common.Hash{common.BytesToHash([]byte{0})}})
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Keys for same address but equal returned true")
	}

	comparedMsg = &Message{AccessList: []types.AccessTuple{{Address: common.Address{1}, StorageKeys: []common.Hash{common.BytesToHash([]byte{1})}}}}
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different AccessList but equal returned true")
	}

	comparedMsg.AccessList = msg.AccessList
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages Value are same but equal returned false")
	}
}

func TestMessage_EqualGasFeeCap(t *testing.T) {
	msg := &Message{GasFeeCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasFeeCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasFeeCap are different but equal returned true")
	}

	comparedMsg.GasFeeCap = msg.GasFeeCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasFeeCap are same but equal returned false")
	}
}

func TestMessage_EqualGasTipCap(t *testing.T) {
	msg := &Message{GasTipCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{GasTipCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages GasTipCap are different but equal returned true")
	}

	comparedMsg.GasTipCap = msg.GasTipCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages GasTipCap are same but equal returned false")
	}
}

func TestMessage_DataHashReturnsIfExists(t *testing.T) {
	want := common.BytesToHash([]byte{1})
	msg := &Message{dataHash: &want}

	got := msg.DataHash()
	if want != got {
		t.Fatalf("hashes are different\nwant: %v\ngot: %v", want, got)
	}

}

func TestMessage_DataHashGeneratesNewHashIfNil(t *testing.T) {
	msg := &Message{Data: []byte{1}}
	got := msg.DataHash()

	want := crypto.Keccak256Hash(msg.Data)

	if got == common.EmptyHash {
		t.Fatal("dataHash is nil")
	}

	if want != got {
		t.Fatalf("hashes are different\nwant: %v\ngot: %v", want, got)
	}

}
