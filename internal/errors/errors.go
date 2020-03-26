package errors

// CustomError to Expose additional information for error.
type CustomError interface {
	GetCode() uint32
}

type baseError struct {
	code    uint32
	message string
	inner   error
}

func (e *baseError) Error() string {
	return e.message
}

func (e *baseError) GetCode() uint32 {
	return e.code
}

// New ...
func New(code uint32, message string) error {
	return new(nil, code, message)
}

func new(err error, code uint32, message string) *baseError {
	return &baseError{
		message: message,
		code:    code,
		inner:   err,
	}
}

//Wrap ...
func Wrap(err error, code uint32, message string) error {
	return new(err, code, message)
}
