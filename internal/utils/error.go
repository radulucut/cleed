package utils

type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

func NewInternalError(message string) *InternalError {
	return &InternalError{
		Message: message,
	}
}
