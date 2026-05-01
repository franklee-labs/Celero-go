package condition

// InvalidNodeError is returned when error occurred during condition build.
type InvalidNodeError struct {
	Msg string
}

func (e *InvalidNodeError) Error() string { return e.Msg }

// EvalError wraps an unexpected CEL evaluation failure.
type EvalError struct {
	Cause error
}

func (e *EvalError) Error() string { return e.Cause.Error() }
func (e *EvalError) Unwrap() error { return e.Cause }
