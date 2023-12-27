package new_substate

import (
	"errors"
	"fmt"
)

func NewSubstate(input Alloc, output Alloc, env *Env, message *Message, result *Result) *Substate {
	return &Substate{
		InputAlloc:  input,
		OutputAlloc: output,
		Env:         env,
		Message:     message,
		Result:      result,
	}
}

type Substate struct {
	InputAlloc  Alloc
	OutputAlloc Alloc
	Env         *Env
	Message     *Message
	Result      *Result
}

// Equal returns true if s is y or if values of s are equal to values of y.
// Otherwise, s and y are not equal hence false is returned.
func (s *Substate) Equal(y *Substate) (err error) {
	if s == y {
		return nil
	}

	if (s == nil || y == nil) && s != y {
		return errors.New("one of the substates is nil")
	}

	input := s.InputAlloc.Equal(y.InputAlloc)
	output := s.OutputAlloc.Equal(y.OutputAlloc)
	env := s.Env.Equal(y.Env)
	msg := s.Message.Equal(y.Message)
	res := s.Result.Equal(y.Result)

	if !input {
		err = errors.Join(err, fmt.Errorf("input alloc is different\nwant: %v\n got: %v", s.InputAlloc.String(), y.InputAlloc.String()))
	}

	if !output {
		err = errors.Join(err, fmt.Errorf("output alloc is different\nwant: %v\n got: %v", s.OutputAlloc.String(), y.OutputAlloc.String()))
	}

	if !env {
		err = errors.Join(err, fmt.Errorf("env is different\nwant: %v\n got: %v", s.Env.String(), y.Env.String()))
	}

	if !msg {
		err = errors.Join(err, fmt.Errorf("message is different\nwant: %v\n got: %v", s.Message.String(), y.Message.String()))
	}

	if !res {
		err = errors.Join(err, fmt.Errorf("result is different\nwant: %v\n got: %v", s.Result.String(), y.Result.String()))
	}

	return err
}
