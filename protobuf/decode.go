package protobuf

import (
	"github.com/Fantom-foundation/Substate/types"
	"github.com/Fantom-foundation/Substate/substate"
	pb "github.com/Fantom-foundation/Substate/protobuf"
)

// Decode converts protobuf-encoded Substate into aida-comprehensible substate
func (s *pb.Substate) Decode() (*substate.Substate, error) {
	input, err := s.GetInputAlloc().decode()
	if err != nil {
		return nil, err
	}

	output, err := s.GetInputAlloc().decode()
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
	world := make(substate.WorldState)

	for _, entry := range Substate_AllocEntry {
		addr := types.BytesToAddress(entry.Address)
		acct := entry.Account

	}

	return nil, fmt.Errorf("Not Implemented")
}

// decode converts protobuf-encoded Substate_BlockEnv into aida-comprehensible Env
func (env *pb.Substate_BlockEnv) decode() (*substate.Env, error) {
	return nil, fmt.Errorf("Not Implemented")
}

// decode converts protobuf-encoded Substate_TxMessage into aida-comprehensible Message
func (msg *pb.Substate_TxMessage) decode() (*substate.Message, error) {
	/*return &Message{
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
	return nil, fmt.Errorf("Not Implemented")
}

