package iterator

type Iterator[T comparable] interface {
	// Next moves the iterator to the next key/value pair. It returns whether the
	// iterator is exhausted.
	Next() bool

	// Error returns any accumulated error. Exhausting all the key/value pairs
	// is not considered to be an error.
	Error() error

	// Value returns the current value of type T, or nil if done. The
	// caller should not modify the contents of the returned slice, and its contents
	// may change on the next call to Next.
	Value() T

	// Release releases associated resources. Release should always succeed and can
	// be called multiple times without causing error.
	Release()
}
