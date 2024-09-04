package protobuf

import (
	"github.com/Fantom-foundation/Substate/substate"
	"github.com/Fantom-foundation/Substate/types"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// Decode converts protobuf-encoded Substate into aida-comprehensible substate
func (s *Substate) Decode(block uint64, tx int) (*substate.Substate, error) {
	input, err := s.GetInputAlloc().decode()
	if err != nil {
		return nil, err
	}

	output, err := s.GetOutputAlloc().decode()
	if err != nil {
		return nil, err
	}

	environment, err := s.GetBlockEnv().decode()
	if err != nil {
		return nil, err
	}

	message, err := s.GetTxMessage().decode()
	if err != nil {
		return nil, err
	}

	result, err := s.GetResult().decode()
	if err != nil {
		return nil, err
	}

	return &substate.Substate{
		InputSubstate:  input,
		OutputSubstate: output,
		Env:            environment,
		Message:        message,
		Result:         result,
		Block:          block,
		Transaction:    tx,
	}, nil
}

// decode converts protobuf-encoded Substate_Alloc into aida-comprehensible WorldState
func (alloc *Substate_Alloc) decode() (*substate.WorldState, error) {
	world := make(substate.WorldState, len(alloc.GetAlloc()))

	for _, entry := range alloc.GetAlloc() {
		addr, acct, err := entry.decode()
		if err != nil {
			return nil, fmt.Errorf("Error decoding alloc entry; %w", err)
		}

		address := types.BytesToAddress(addr)
		nonce, balance, codehash, err := acct.decode()
		if err != nil {
			return nil, fmt.Errorf("Error decoding entry account; %w", err)
		}

		world = world.Add(address, nonce, balance, codehash)
	}

	return world, nil
}

func (entry *Substate_AllocEntry) decode() ([]byte, *pb.Substate_Account, error) {
	return entry.GetAddress(), entry.GetAccount(), nil
}

func (acct *Substate_Account) decode() (uint64, *big.Int, []byte, error) {
	return acct.GetNonce(),
		new(big.Int).SetBytes(acct.GetBalance()),
		acct.GetCodeHash(),
		nil
}

// decode converts protobuf-encoded Substate_BlockEnv into aida-comprehensible Env
func (env *Substate_BlockEnv) decode() (*substate.Env, error) {
	blockHashes := make(map[uint64]types.Hash, len(env.GetBlockHashes()))
	for _, entry := range env.GetBlockHashes() {
		key, value, err := entry.decode()
		if err != nil {
			return nil, err
		}
		blockHashes[key] := types.BytesToHash(value)
	}

	var baseFee *big.Int = nil
	if env.GetBaseFee() != nil {
		baseFee = &new(big.Int).SetBytes(env.GetBaseFee().GetValue())
	}

	var blobBaseFee *big.Int = nil
	if env.GetBlobBaseFee() != nil {
		blobBaseFee = &new(big.Int).SetBytes(env.GetBlobBaseFee().GetValue())
	}

	return &substate.Env{
		Coinbase:    types.BytesToAddress(env.GetCoinbase()),
		Difficulty:  new(big.Int).SetBytes(env.GetDifficulty()),
		GasLimit:    env.GetGasLimit(),
		Number:      env.GetNumber(),
		Timestamp:   env.GetTimestamp(),
		BlockHashes: blockHashes,
		BaseFee:     baseFee,
		//Random: env.GetRandom(), // does not exist in substate.Env
		BlobBaseFee: blobBaseFee,
	}, nil
}

func (entry *Substate_BlockEnv_BlockHashEntry) decode() (uint64, []byte, error) {
	return entry.GetKey(), entry.GetValue(), nil
}

// decode converts protobuf-encoded Substate_TxMessage into aida-comprehensible Message
func (msg *Substate_TxMessage) decode() (*substate.Message, error) {
	// to=nil means contract creation
	var toAddr *types.Address = nil
	to := msg.GetTo()
	if to != nil {
		toAddr = &types.BytesToAddress(to.GetValue())
	}

	// Berlin hard fork, EIP-2930: Optional access lists
	var accessList types.AccessList = nil // nil if EIP-2930 is not activated
	if msg.GetAccessList() != nil {
		accessList = make([]AccessTuple, len(msg.GetAccessList()))
		for i, entry := range msg.GetAccessList() {
			addr, keys, err := entry.decode()
			if err != nil {
				return nil, err
			}

			address := types.BytesToAddress(addr)
			storageKeys := make([]types.Hash, len(keys))
			for j, key := range keys {
				storageKeys[j] := types.BytesToHash(key)
			}

			accessList[i] := &AccessTuple{
				Address:     address,
				StorageKeys: storageKeys,
			}
		}
	}

	// London hard fork, EIP-1559: Fee market
	var gasFeeCap *big.Int = nil
	gfc := msg.GetGasFeeCap()
	if gfc != nil {
		gasFeeCap = &new(big.Int).SetBytes(gfc.GetValue())
	}

	var gasTipCap *big.Int = nil
	gtc := msg.GetGasTipCap()
	if gtc != nil {
		gasTipCap = &new(big.Int).SetBytes(gtc.GetValue())
	}

	// Cancun hard fork, EIP-4844
	var blobGasFeeCap *big.Int = nil
	bgfc := msg.GetBlobGasFeeCap()
	if bgfc != nil {
		blobGasFeeCap = &new(big.Int).SetBytes(bgfc.GetValue())
	}

	blobHashes := make([]types.Hash, len(msg.GetBlobHashes()))
	for i, hash := range msg.GetBlobHashes() {
		blobHashes[i] := types.BytesToHash(hash)
	}

	return &substate.Message{
		Nonce:         msg.GetNonce(),
		CheckNonce:    true, //always true
		GasPrice:      new(big.Int).SetBytes(msg.GetGasPrice()),
		Gas:           msg.GetGas(),
		From:          types.BytesToAddress(msg.GetFrom()),
		To:            toAddr,
		Value:         new(big.Int).SetBytes(msg.GetValue()),
		Data:          msg.GetData(),
		dataHash:      msg.GetInitCodeHash(),
		AccessList:    accessList,
		GasFeeCap:     gasFeeCap,
		GasTipCap:     gasTipCap,
		BlobGasFeeCap: blobGasFeeCap,
		BlobHashes:    blobHashes,
	}, nil
}

func (entry *Substate_TxMessage_AccessListEntry) decode() ([]byte, [][]byte, error) {
	return entry.GetAddress(), entry.GetStorageKeys(), nil
}

// decode converts protobuf-encoded Substate_Result into aida-comprehensible Result
func (res *Substate_Result) decode() (*substate.Result, error) {
	logs := make([]types.Log, len(res.GetLogs))
	for i, log := range res.GetLogs() {
		logs[i] := log.decode()
	}

	return &substate.Result{
		Status:          &res.Status,
		Bloom:           types.BytesToBloom(res.Bloom),
		Logs:            logs,
		ContractAddress: nil, // to be processed downstream
		GasUsed:         &res.GasUsed,
	}, nil
}

func (log *Substate_Result_Log) decode() (*types.Log, error) {
	topics := make([]types.Hash, len(log.GetTopics()))
	for i, topic := range log.GetTopics() {
		topics[i] := types.BytesToHash(topic)
	}

	return &types.Log{
		Address: types.BytesToAddress(log.GetAddress()),
		Topics:  topics,
		Data:    log.GetData(),
	}, nil
}
