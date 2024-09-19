package protobuf

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/syndtr/goleveldb/leveldb"
)

type dbGetCode = func(types.Hash) ([]byte, error)

// Decode converts protobuf-encoded Substate into aida-comprehensible substate
func (s *Substate) Decode(lookup dbGetCode, block uint64, tx int) (*substate.Substate, error) {
	input, err := s.GetInputAlloc().decode(lookup)
	if err != nil {
		return nil, err
	}

	output, err := s.GetOutputAlloc().decode(lookup)
	if err != nil {
		return nil, err
	}

	environment, err := s.GetBlockEnv().decode()
	if err != nil {
		return nil, err
	}

	message, err := s.GetTxMessage().decode(lookup)
	if err != nil {
		return nil, err
	}

	contractAddress := s.GetTxMessage().getContractAddress()
	result, err := s.GetResult().decode(contractAddress)
	if err != nil {
		return nil, err
	}

	return &substate.Substate{
		InputSubstate:  *input,
		OutputSubstate: *output,
		Env:            environment,
		Message:        message,
		Result:         result,
		Block:          block,
		Transaction:    tx,
	}, nil
}

// decode converts protobuf-encoded Substate_Alloc into aida-comprehensible WorldState
func (alloc *Substate_Alloc) decode(lookup dbGetCode) (*substate.WorldState, error) {
	world := make(substate.WorldState, len(alloc.GetAlloc()))

	for _, entry := range alloc.GetAlloc() {
		addr, acct, err := entry.decode()
		if err != nil {
			return nil, fmt.Errorf("Error decoding alloc entry; %w", err)
		}

		address := types.BytesToAddress(addr)
		nonce, balance, _, codehash, err := acct.decode()
		if err != nil {
			return nil, fmt.Errorf("Error decoding entry account; %w", err)
		}

		code, err := lookup(codehash)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("Error looking up codehash; %w", err)
		}

		world[address] = substate.NewAccount(nonce, balance, code)
		for _, storage := range acct.GetStorage() {
			key, value, err := storage.decode()
			if err != nil {
				return nil, fmt.Errorf("Error decoding account storage entry; %w", err)
			}
			world[address].Storage[key] = value
		}
	}

	return &world, nil
}

func (entry *Substate_AllocEntry) decode() ([]byte, *Substate_Account, error) {
	return entry.GetAddress(), entry.GetAccount(), nil
}

func (acct *Substate_Account) decode() (uint64, *uint256.Int, []byte, types.Hash, error) {
	return acct.GetNonce(),
		types.BytesToUint256(acct.GetBalance()),
		acct.GetCode(),
		types.BytesToHash(acct.GetCodeHash()),
		nil
}

func (entry *Substate_Account_StorageEntry) decode() (types.Hash, types.Hash, error) {
	return types.BytesToHash(entry.GetKey()),
		types.BytesToHash(entry.GetValue()),
		nil
}

// decode converts protobuf-encoded Substate_BlockEnv into aida-comprehensible Env
func (env *Substate_BlockEnv) decode() (*substate.Env, error) {
	var blockHashes map[uint64]types.Hash = nil
	if env.GetBlockHashes() != nil {
		blockHashes = make(map[uint64]types.Hash, len(env.GetBlockHashes()))
		for _, entry := range env.GetBlockHashes() {
			key, value, err := entry.decode()
			if err != nil {
				return nil, err
			}
			blockHashes[key] = types.BytesToHash(value)
		}
	}

	return &substate.Env{
		Coinbase:    types.BytesToAddress(env.GetCoinbase()),
		Difficulty:  types.BytesToBigInt(env.GetDifficulty()),
		GasLimit:    env.GetGasLimit(),
		Number:      env.GetNumber(),
		Timestamp:   env.GetTimestamp(),
		BlockHashes: blockHashes,
		BaseFee:     BytesValueToBigInt(env.GetBaseFee()),
		Random:      BytesValueToHash(env.GetRandom()),
		BlobBaseFee: BytesValueToBigInt(env.GetBlobBaseFee()),
	}, nil
}

func (entry *Substate_BlockEnv_BlockHashEntry) decode() (uint64, []byte, error) {
	return entry.GetKey(), entry.GetValue(), nil
}

// decode converts protobuf-encoded Substate_TxMessage into aida-comprehensible Message
func (msg *Substate_TxMessage) decode(lookup dbGetCode) (*substate.Message, error) {

	// to=nil means contract creation
	var pTo *types.Address = nil
	to := msg.GetTo()
	if to != nil {
		address := types.BytesToAddress(to.GetValue())
		pTo = &address
	}

	// if InitCodeHash exists:
	// 1. code = lookup the code using InitCodeHash
	// 2. set data -> code from (1)
	// 3. clear InitCodeHash
	var data []byte = msg.GetData()
	if pTo == nil {
		code, err := lookup(types.BytesToHash(msg.GetInitCodeHash()))
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("failed to decode tx message; %w", err)
		}
		data = code
	}

	txType := msg.GetTxType()

	// Berlin hard fork, EIP-2930: Optional access lists
	var accessList types.AccessList = nil // nil if EIP-2930 is not activated
	switch txType {
	case 	Substate_TxMessage_TXTYPE_ACCESSLIST,
		Substate_TxMessage_TXTYPE_DYNAMICFEE,
		Substate_TxMessage_TXTYPE_BLOB:

		accessList = make([]types.AccessTuple, len(msg.GetAccessList()))
		for i, entry := range msg.GetAccessList() {
			addr, keys, err := entry.decode()
			if err != nil {
				return nil, err
			}

			address := types.BytesToAddress(addr)
			storageKeys := make([]types.Hash, len(keys))
			for j, key := range keys {
				storageKeys[j] = types.BytesToHash(key)
			}

			accessList[i] = types.AccessTuple{
				Address:     address,
				StorageKeys: storageKeys,
			}
		}
	}

	// London hard fork, EIP-1559: Fee market
	var gasFeeCap *big.Int = types.BytesToBigInt(msg.GetGasPrice())
	var gasTipCap *big.Int = types.BytesToBigInt(msg.GetGasPrice())
	switch txType {
	case 	Substate_TxMessage_TXTYPE_DYNAMICFEE, 
		Substate_TxMessage_TXTYPE_BLOB:

		gasFeeCap = BytesValueToBigInt(msg.GetGasFeeCap())
		gasTipCap = BytesValueToBigInt(msg.GetGasTipCap())
	}

	// Cancun hard fork, EIP-4844
	var blobHashes []types.Hash = nil
	switch txType {
	case Substate_TxMessage_TXTYPE_BLOB:
		if msg.GetBlobHashes() != nil {
			blobHashes = make([]types.Hash, len(msg.GetBlobHashes()))
			for i, hash := range msg.GetBlobHashes() {
				blobHashes[i] = types.BytesToHash(hash)
			}
		}
	}

	return &substate.Message{
		Nonce:         msg.GetNonce(),
		CheckNonce:    true,
		GasPrice:      types.BytesToBigInt(msg.GetGasPrice()),
		Gas:           msg.GetGas(),
		From:          types.BytesToAddress(msg.GetFrom()),
		To:            pTo,
		Value:         types.BytesToBigInt(msg.GetValue()),
		Data:          data,
		AccessList:    accessList,
		GasFeeCap:     gasFeeCap,
		GasTipCap:     gasTipCap,
		BlobGasFeeCap: BytesValueToBigInt(msg.GetBlobGasFeeCap()),
		BlobHashes:    blobHashes,
	}, nil
}

func (entry *Substate_TxMessage_AccessListEntry) decode() ([]byte, [][]byte, error) {
	return entry.GetAddress(), entry.GetStorageKeys(), nil
}

// getContractAddress returns the address of the newly created contract if any.
// returns nil otherwise.
func (msg *Substate_TxMessage) getContractAddress() common.Address {
	var contractAddress common.Address

	// *to==nil means contract creation and thus address of newly created contract
	to := msg.GetTo()
	if to == nil {
		fromAddr := common.BytesToAddress(msg.GetFrom())
		contractAddress = crypto.CreateAddress(fromAddr, msg.GetNonce())
	}

	return contractAddress
}

// decode converts protobuf-encoded Substate_Result into aida-comprehensible Result
func (res *Substate_Result) decode(contractAddress common.Address) (*substate.Result, error) {
	var err error = nil
	logs := make([]*types.Log, len(res.GetLogs()))
	for i, log := range res.GetLogs() {
		logs[i], err = log.decode()
		if err != nil {
			return nil, fmt.Errorf("Error decoding result; %w", err)
		}
	}

	return substate.NewResult(
		res.GetStatus(),               // Status
		types.BytesToBloom(res.Bloom), // Bloom
		logs,                          // Logs
		types.BytesToAddress(contractAddress.Bytes()), // ContractAddress
		res.GetGasUsed(), // GasUsed
	), nil
}

func (log *Substate_Result_Log) decode() (*types.Log, error) {
	topics := make([]types.Hash, len(log.GetTopics()))
	for i, topic := range log.GetTopics() {
		topics[i] = types.BytesToHash(topic)
	}

	return &types.Log{
		Address: types.BytesToAddress(log.GetAddress()),
		Topics:  topics,
		Data:    log.GetData(),
	}, nil
}
