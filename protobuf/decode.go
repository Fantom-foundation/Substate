package protobuf

import (
	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/substate"
	pb "github.com/Fantom-foundation/Substate/protobuf"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// Decode converts protobuf-encoded Substate into aida-comprehensible substate
func (s *pb.Substate) Decode() (*substate.Substate, error) {
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
func (alloc *pb.Substate_Alloc) decode() (*substate.WorldState, error) {
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

func (entry *pb.Substate_AllocEntry) decode() ([]byte, *pb.Substate_Account, error) {
	return entry.GetAddress(), entry.GetAccount(), nil
}

func (acct *pb.Substate_Account) decode() (uint64, *big.Int, []byte, error) {
	return 	acct.GetNonce(),
		new(big.Int).SetBytes(acct.GetBalance()),
		acct.GetCodeHash(),
		nil
}

// decode converts protobuf-encoded Substate_BlockEnv into aida-comprehensible Env
func (env *pb.Substate_BlockEnv) decode() (*substate.Env, error) {
	blockHashes := make(map[uint64]types.Hash, len(env.GetBlockHashes()))
	for _, entry := range env.GetBlockHashes() {
		key, value, err := entry.decode()
		if err != nil {
			return nil, err
		}
		blockHashes[key] := types.BytesToHash(value)
	}

	return &substate.Env{
		Coinbase: types.BytesToAddress(env.GetCoinbase()),
		Difficulty: new(big.Int).SetBytes(env.GetDifficulty()),
		GasLimit: env.GetGasLimit(),
		Number: env.GetNumber(),
		Timestamp: env.GetTimestamp(),
		BlockHashes: blockHashes,
		BaseFee := new(big.Int).SetBytes(env.GetBaseFee().GetValue()),
		//Random := env.GetRandom(), // does not exist in substate.Env
		BlobBaseFee := new(big.Int).SetBytes(env.GetBlobBaseFee().GetValue()),
	}, nil
}

func (entry *pb.Substate_BlockEnv_BlockHashEntry) decode() (uint64, []byte, error) {
	return entry.GetKey(), entry.GetValue(), nil
}


// decode converts protobuf-encoded Substate_TxMessage into aida-comprehensible Message
func (msg *pb.Substate_TxMessage) decode() (*substate.Message, error) {
	/*return &substate.Message{
		Nonce:         &msg.Nonce,
		CheckNonce:    nil,
		GasPrice:      nil,
		Gas:           &msg.Gas,
		From:          from,
		To:            to,
		Value:         value,
		Data:          data,
		dataHash:      dataHash,
		AccessList:    accessList,
		GasFeeCap:     gasFeeCap,
		GasTipCap:     gasTipCap,
		BlobGasFeeCap: blobGasFeeCap,
		BlobHashes:    blobHashes,
	}, nil */
	return nil, fmt.Errorf("Not Implemented")
}

// decode converts protobuf-encoded Substate_Result into aida-comprehensible Result
func (res *pb.Substate_Result) decode() (*substate.Result, error) {
	logs := make([]types.Log, len(res.Logs))
	for i, log := res.Logs {
		logs[i] := log.decode()
	}

	return &substate.Result{
		Status: &res.Status,
		Bloom: types.BytesToBloom(res.Bloom),
		Logs: logs,
		ContractAddress: nil, // to be processed downstream
		GasUsed: &res.GasUsed,
	}, nil
}

func (log *pb.Substate_Result_Log) decode() (*types.Log, error) {
	topics := make([]types.Hash, len(log.GetTopics()))
	for i, topic := range log.GetTopics() {
		topics[i] := types.BytesToHash(topic)
	}

	return &types.Log{
		Address: types.BytesToAddress(log.GetAddress()),
		Topics:  topics,
		Data: log.GetData(),
	}, nil
}
