package handler

import (
	cep "climaCepOtel/internal/domain/vo"
	"climaCepOtel/internal/usecase"
	"climaCepOtel/pkgs/apperror"
	"context"
	"errors"
	"net/http"
)

// na borda HTTP (handler/adapter), traduz erro de domínio em AppError
func toAppError(err error) *apperror.AppError {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &apperror.AppError{Code: http.StatusRequestTimeout, Message: "request timeout", Err: err}
	case errors.Is(err, cep.ErrorInvalidCep):
		return apperror.NewUnprocessable(err)
	case errors.Is(err, usecase.ErrorCepNotFound):
		return apperror.NewNotFound(err)
	default:
		return apperror.NewInternal(err)
	}
}

func errorHandler(err error, w http.ResponseWriter) {
	status := http.StatusInternalServerError
	message := "internal server error"

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		status = appErr.StatusCode()
		message = appErr.Error()
	}

	http.Error(w, message, status)
}
