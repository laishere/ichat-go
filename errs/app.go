package errs

type AppError interface {
	Code() int
	Error() string
}

type appError struct {
	code  int
	error string
}

func (e appError) Code() int {
	return e.code
}

func (e appError) Error() string {
	return e.error
}

func NewAppError(code int, error string) AppError {
	return appError{code, error}
}
