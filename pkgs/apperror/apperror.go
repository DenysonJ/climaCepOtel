package apperror

import "net/http"

type AppError struct {
	Code    int    // HTTP status
	Message string // mensagem segura para o cliente
	Err     error  // erro original (domínio), para log e errors.Is/As
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) StatusCode() int {
	return e.Code
}

// Unwrap permite que errors.Is/As continuem enxergando o erro de domínio embrulhado.
func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFound(err error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: err.Error(), Err: err}
}

func NewUnprocessable(err error) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: err.Error(), Err: err}
}

func NewInternal(err error) *AppError {
	// mensagem genérica — não vaza detalhe interno
	return &AppError{Code: http.StatusInternalServerError, Message: "internal server error", Err: err}
}
