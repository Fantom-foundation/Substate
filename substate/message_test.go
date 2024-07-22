package substate

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/types/hash"
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
	msg := &Message{From: types.Address{0}}
	comparedMsg := &Message{From: types.Address{1}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages From are different but equal returned true")
	}

	comparedMsg.From = msg.From
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages From are same but equal returned false")
	}
}

func TestMessage_EqualTo(t *testing.T) {
	msg := &Message{To: &types.Address{0}}
	comparedMsg := &Message{To: &types.Address{1}}

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
	msg := &Message{dataHash: new(types.Hash)}
	*msg.dataHash = types.BytesToHash([]byte{0})
	comparedMsg := &Message{dataHash: new(types.Hash)}
	*comparedMsg.dataHash = types.BytesToHash([]byte{1})

	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal even if it is different")
	}

	comparedMsg.dataHash = msg.dataHash
	if !msg.Equal(comparedMsg) {
		t.Fatal("dataHash must not affect equal")
	}
}

func TestMessage_EqualAccessList(t *testing.T) {
	msg := &Message{AccessList: []types.AccessTuple{{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{0})}}}}
	comparedMsg := &Message{AccessList: []types.AccessTuple{{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Key for same address but equal returned true")
	}

	comparedMsg.AccessList = append(comparedMsg.AccessList, types.AccessTuple{Address: types.Address{0}, StorageKeys: []types.Hash{types.BytesToHash([]byte{0})}})
	if msg.Equal(comparedMsg) {
		t.Fatal("messages access list have different Storage Keys for same address but equal returned true")
	}

	comparedMsg = &Message{AccessList: []types.AccessTuple{{Address: types.Address{1}, StorageKeys: []types.Hash{types.BytesToHash([]byte{1})}}}}
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

func TestMessage_EqualBlobGasFeeCap(t *testing.T) {
	msg := &Message{BlobGasFeeCap: new(big.Int).SetUint64(0)}
	comparedMsg := &Message{BlobGasFeeCap: new(big.Int).SetUint64(1)}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages BlobGasFeeCap are different but equal returned true")
	}

	comparedMsg.BlobGasFeeCap = msg.BlobGasFeeCap
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages BlobGasFeeCap are same but equal returned false")
	}
}

func TestMessage_EqualBlobHashes(t *testing.T) {
	msg := &Message{BlobHashes: []types.Hash{types.BytesToHash([]byte{0x0})}}
	comparedMsg := &Message{BlobHashes: []types.Hash{types.BytesToHash([]byte{0x1})}}

	if msg.Equal(comparedMsg) {
		t.Fatal("messages BlobHashes are different but equal returned true")
	}

	comparedMsg.BlobHashes = msg.BlobHashes
	if !msg.Equal(comparedMsg) {
		t.Fatal("messages BlobHashes are same but equal returned false")
	}
}

func TestMessage_DataHashReturnsIfExists(t *testing.T) {
	want := types.BytesToHash([]byte{1})
	msg := &Message{dataHash: &want}

	got := msg.DataHash()
	if want != got {
		t.Fatalf("hashes are different\nwant: %v\ngot: %v", want, got)
	}

}

func TestMessage_DataHashGeneratesNewHashIfNil(t *testing.T) {
	msg := &Message{Data: []byte{1}}
	got := msg.DataHash()

	want := hash.Keccak256Hash(msg.Data)

	if got.IsEmpty() {
		t.Fatal("dataHash is nil")
	}

	if want != got {
		t.Fatalf("hashes are different\nwant: %v\ngot: %v", want, got)
	}

}
