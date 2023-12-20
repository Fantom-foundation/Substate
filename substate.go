package substate

// todo use this substate once ready
type substate struct {
	InputAlloc  Alloc
	OutputAlloc Alloc
	Env         *Env
	Message     *Message
	Result      *Result
}

// Equal returns true if s is y or if values of s are equal to values of y.
// Otherwise, s and y are not equal hence false is returned.
func (s *substate) Equal(y *substate) bool {
	if s == y {
		return true
	}

	if (s == nil || y == nil) && s != y {
		return false
	}

	return s.InputAlloc.Equal(y.InputAlloc) &&
		s.OutputAlloc.Equal(y.OutputAlloc) &&
		s.Env.Equal(y.Env) &&
		s.Message.Equal(y.Message) &&
		s.Result.Equal(y.Result)
}
