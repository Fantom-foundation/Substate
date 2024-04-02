package substate

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// SubstateAccount is modification of GenesisAccount in core/genesis.go
// Deprecated: This type is getting deleted, use substate.Account instead
type SubstateAccount struct {
	Nonce   uint64
	Balance *big.Int
	Storage map[common.Hash]common.Hash
	Code    []byte
}

// Deprecated: This type is getting deleted, use substate.NewAccount instead
func NewSubstateAccount(nonce uint64, balance *big.Int, code []byte) *SubstateAccount {
	return &SubstateAccount{
		Nonce:   nonce,
		Balance: new(big.Int).Set(balance),
		Storage: make(map[common.Hash]common.Hash),
		Code:    code,
	}
}

// Deprecated: Use NewAccount
// This type is getting deleted, please check Account
func (x *SubstateAccount) Equal(y *SubstateAccount) bool {
	if x == y {
		return true
	}

	if (x == nil || y == nil) && x != y {
		return false
	}

	equal := (x.Nonce == y.Nonce &&
		x.Balance.Cmp(y.Balance) == 0 &&
		bytes.Equal(x.Code, y.Code) &&
		len(x.Storage) == len(y.Storage))
	if !equal {
		return false
	}

	for k, xv := range x.Storage {
		yv, exist := y.Storage[k]
		if !(exist && xv == yv) {
			return false
		}
	}

	return true
}

// Deprecated: Use NewAccount
// This type is getting deleted, please check Account
func (sa *SubstateAccount) Copy() *SubstateAccount {
	saCopy := NewSubstateAccount(sa.Nonce, sa.Balance, sa.Code)

	for key, value := range sa.Storage {
		saCopy.Storage[key] = value
	}

	return saCopy
}

func (sa *SubstateAccount) CodeHash() common.Hash {
	return crypto.Keccak256Hash(sa.Code)
}

// Deprecated: Use substate.Alloc
type SubstateAlloc map[common.Address]*SubstateAccount

// EstinateIncrementalSize returns estimated substate size increase after merge
func (x SubstateAlloc) EstimateIncrementalSize(y SubstateAlloc) uint64 {
	var (
		size          uint64 = 0
		sizeOfAddress uint64 = 20
		sizeOfHash    uint64 = 32
		sizeOfNonce   uint64 = 8
	)
	for addr, account := range y {
		if xaccount, found := x[addr]; found {
			// skip if no diff
			if xaccount.Equal(account) {
				continue
			}
			// update storage by y
			for key, _ := range account.Storage {
				// only add new storage keys
				if _, found := x[addr].Storage[key]; !found {
					size += sizeOfHash // add sizeof(common.Hash)
				}
			}
		} else {
			// add size of new accounts
			// address + nonce + balance + codehash
			size += sizeOfAddress + sizeOfNonce + uint64(len(account.Balance.Bytes())) + sizeOfHash
			// storage slots * sizeof(common.Hash)
			size += uint64(len(account.Storage)) * sizeOfHash
		}
	}
	return size
}

func (x SubstateAlloc) Merge(y SubstateAlloc) {
	for addr, account := range y {
		if xaccount, found := x[addr]; found {
			if xaccount.Equal(account) {
				continue
			}
			// overwrite account details in x by y
			x[addr].Nonce = account.Nonce
			x[addr].Balance = new(big.Int).Set(account.Balance)
			x[addr].Code = make([]byte, len(account.Code))
			copy(x[addr].Code, account.Code)
		} else {
			// create new account details in x
			x[addr] = NewSubstateAccount(account.Nonce, account.Balance, account.Code)
		}
		// update storage by y
		for key, value := range account.Storage {
			x[addr].Storage[key] = value
		}
	}
}

// Diff computes the difference set between two substate alloc (z = x\y).
// Note: Zero value and non-existing value are considered equal.
func (x SubstateAlloc) Diff(y SubstateAlloc) SubstateAlloc {
	z := make(SubstateAlloc)
	for addr, account := range x {
		if yaccount, found := y[addr]; !found {
			z[addr] = account.Copy()
		} else {
			if yaccount.Equal(account) {
				continue
			} else {
				// check nonce, balance and code
				equal := (account.Nonce == yaccount.Nonce &&
					account.Balance.Cmp(yaccount.Balance) == 0 &&
					bytes.Equal(account.Code, yaccount.Code))
				if !equal {
					z[addr] = NewSubstateAccount(account.Nonce, account.Balance, account.Code)
				}

				// check storage
				for key, value := range account.Storage {
					if yvalue, found := y[addr].Storage[key]; (!found && value != common.Hash{}) || yvalue != value {
						// initialize if not exists.
						if _, found := z[addr]; !found {
							z[addr] = NewSubstateAccount(account.Nonce, account.Balance, account.Code)
						}
						z[addr].Storage[key] = value
					}
				}
			}
		}
	}
	return z
}

func (x SubstateAlloc) Equal(y SubstateAlloc) bool {
	if len(x) != len(y) {
		return false
	}

	for k, xv := range x {
		yv, exist := y[k]
		if !(exist && xv.Equal(yv)) {
			return false
		}
	}

	return true
}

type SubstateEnv struct {
	Coinbase    common.Address
	Difficulty  *big.Int
	GasLimit    uint64
	Number      uint64
	Timestamp   uint64
	BlockHashes map[uint64]common.Hash

	// London hard fork, EIP-1559
	BaseFee *big.Int // nil if EIP-1559 is not activated
}

func NewSubstateEnv(b *types.Block, blockHashes map[uint64]common.Hash) *SubstateEnv {
	var env = &SubstateEnv{}

	env.Coinbase = b.Coinbase()
	env.Difficulty = new(big.Int).Set(b.Difficulty())
	env.GasLimit = b.GasLimit()
	env.Number = b.NumberU64()
	env.Timestamp = b.Time()
	env.BlockHashes = make(map[uint64]common.Hash)
	for num64, bhash := range blockHashes {
		env.BlockHashes[num64] = bhash
	}

	env.BaseFee = b.BaseFee()

	return env
}

func (x *SubstateEnv) Equal(y *SubstateEnv) bool {
	if x == y {
		return true
	}

	if (x == nil || y == nil) && x != y {
		return false
	}

	equal := (x.Coinbase == y.Coinbase &&
		x.Difficulty.Cmp(y.Difficulty) == 0 &&
		x.GasLimit == y.GasLimit &&
		x.Number == y.Number &&
		x.Timestamp == y.Timestamp &&
		len(x.BlockHashes) == len(y.BlockHashes) &&
		x.BaseFee.Cmp(y.BaseFee) == 0)
	if !equal {
		return false
	}

	for k, xv := range x.BlockHashes {
		yv, exist := y.BlockHashes[k]
		if !(exist && xv == yv) {
			return false
		}
	}

	return true
}

type SubstateMessage struct {
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

func NewSubstateMessage(msg *types.Message) *SubstateMessage {
	var smsg = &SubstateMessage{}

	smsg.Nonce = msg.Nonce()
	smsg.CheckNonce = !msg.IsFake()
	smsg.GasPrice = msg.GasPrice()
	smsg.Gas = msg.Gas()

	smsg.From = msg.From()
	smsg.To = msg.To()
	smsg.Value = msg.Value()
	smsg.Data = msg.Data()

	smsg.AccessList = msg.AccessList()

	smsg.GasFeeCap = msg.GasFeeCap()
	smsg.GasTipCap = msg.GasTipCap()

	return smsg
}

func (x *SubstateMessage) Equal(y *SubstateMessage) bool {
	if x == y {
		return true
	}

	if (x == nil || y == nil) && x != y {
		return false
	}

	equal := (x.Nonce == y.Nonce &&
		x.CheckNonce == y.CheckNonce &&
		x.GasPrice.Cmp(y.GasPrice) == 0 &&
		x.Gas == y.Gas &&
		x.From == y.From &&
		(x.To == y.To || (x.To != nil && y.To != nil && *x.To == *y.To)) &&
		x.Value.Cmp(y.Value) == 0 &&
		bytes.Equal(x.Data, y.Data) &&
		len(x.AccessList) == len(y.AccessList) &&
		x.GasFeeCap.Cmp(y.GasFeeCap) == 0 &&
		x.GasTipCap.Cmp(y.GasTipCap) == 0)
	if !equal {
		return false
	}

	for i, xa := range x.AccessList {
		ya := y.AccessList[i]
		equal := (xa.Address == ya.Address &&
			len(xa.StorageKeys) == len(ya.StorageKeys))
		if !equal {
			return false
		}
		for j, xk := range xa.StorageKeys {
			yk := ya.StorageKeys[j]
			if xk != yk {
				return false
			}
		}
	}

	return true
}

func (msg *SubstateMessage) DataHash() common.Hash {
	if msg.dataHash == nil {
		dataHash := crypto.Keccak256Hash(msg.Data)
		msg.dataHash = &dataHash
	}
	return *msg.dataHash
}

func (msg *SubstateMessage) AsMessage() types.Message {
	return types.NewMessage(
		msg.From, msg.To, msg.Nonce, msg.Value,
		msg.Gas, msg.GasPrice, msg.GasFeeCap, msg.GasTipCap,
		msg.Data, msg.AccessList, !msg.CheckNonce)
}

// modification of types.Receipt
type SubstateResult struct {
	Status uint64
	Bloom  types.Bloom
	Logs   []*types.Log

	ContractAddress common.Address
	GasUsed         uint64
}

func NewSubstateResult(receipt *types.Receipt) *SubstateResult {
	var sr = &SubstateResult{}

	sr.Status = receipt.Status
	sr.Bloom = receipt.Bloom
	sr.Logs = receipt.Logs

	sr.ContractAddress = receipt.ContractAddress
	sr.GasUsed = receipt.GasUsed

	return sr
}

func (x *SubstateResult) Equal(y *SubstateResult) bool {
	if x == y {
		return true
	}

	if (x == nil || y == nil) && x != y {
		return false
	}

	equal := (x.Status == y.Status &&
		x.Bloom == y.Bloom &&
		len(x.Logs) == len(y.Logs) &&
		x.ContractAddress == y.ContractAddress &&
		x.GasUsed == y.GasUsed)
	if !equal {
		return false
	}

	for i, xl := range x.Logs {
		yl := y.Logs[i]

		equal := (xl.Address == yl.Address &&
			len(xl.Topics) == len(yl.Topics) &&
			bytes.Equal(xl.Data, yl.Data))
		if !equal {
			return false
		}

		for i, xt := range xl.Topics {
			yt := yl.Topics[i]
			if xt != yt {
				return false
			}
		}
	}

	return true
}

type Substate struct {
	InputAlloc  SubstateAlloc
	OutputAlloc SubstateAlloc
	Env         *SubstateEnv
	Message     *SubstateMessage
	Result      *SubstateResult
}

func NewSubstate(inputAlloc SubstateAlloc, outputAlloc SubstateAlloc, env *SubstateEnv, message *SubstateMessage, result *SubstateResult) *Substate {
	return &Substate{
		InputAlloc:  inputAlloc,
		OutputAlloc: outputAlloc,
		Env:         env,
		Message:     message,
		Result:      result,
	}
}

func (x *Substate) Equal(y *Substate) bool {
	if x == y {
		return true
	}

	if (x == nil || y == nil) && x != y {
		return false
	}

	equal := (x.InputAlloc.Equal(y.InputAlloc) &&
		x.OutputAlloc.Equal(y.OutputAlloc) &&
		x.Env.Equal(y.Env) &&
		x.Message.Equal(y.Message) &&
		x.Result.Equal(y.Result))
	return equal
}
