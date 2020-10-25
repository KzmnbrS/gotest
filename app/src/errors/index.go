package errors

type ValueError struct {
	Description string
}

func NewValueError(description string) *ValueError {
	return &ValueError{description}
}

func (e *ValueError) Error() string {
	return e.Description
}

type IOError struct {
	Operation string
	Target    string
	Details   error
}

func NewIOError(operation, target string, details error) *IOError {
	return &IOError{operation, target, details}
}

func (e *IOError) Error() string {
	return e.Operation + " " + e.Target + ": " + e.Details.Error()
}
