package err

// NervError xxx
type AppError struct {
	errCode int
	msg     string
}

// NewError
func NewAppError(errCode int, msg string) error {
	return &AppError{errCode: errCode, msg: msg}
}

// Error
func (p *AppError) Error() string {
	return p.msg
}
