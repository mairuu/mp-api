package httptransport

type HandlerError struct {
	StatusCode int
	Message    string
	err        error
}

func (h *HandlerError) Status() int {
	return h.StatusCode
}

func (h *HandlerError) Error() string {
	if h.Message == "" {
		return h.err.Error()
	}
	return h.Message
}

func (h *HandlerError) Unwrap() error {
	return h.err
}

func NewHandlerError(statusCode int, message string, err error) *HandlerError {
	return &HandlerError{
		StatusCode: statusCode,
		Message:    message,
		err:        err,
	}
}
