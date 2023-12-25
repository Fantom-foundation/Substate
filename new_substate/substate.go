package new_substate

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
func (s *Substate) Equal(y *Substate) bool {
	if s == y {
		return true
	}

	if (s == nil || y == nil) && s != y {
		return false
	}

	equal := s.InputAlloc.Equal(y.InputAlloc) &&
		s.OutputAlloc.Equal(y.OutputAlloc) &&
		s.Env.Equal(y.Env) &&
		s.Message.Equal(y.Message) &&
		s.Result.Equal(y.Result)
	return equal
}
