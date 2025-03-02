package middlewares

import (
	"encoding/json"
	"net/http"

	"github.com/gopher-fleece/gleece/e2e/assets"
)

func MiddlewareBeforeOperation(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("X-pass-before-operation", "true")

	abortBeforeOperation := r.Header.Get("abort-before-operation")
	if abortBeforeOperation == "true" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "abort-before-operation header is set to true"})
		return false
	}

	return true
}

func MiddlewareAfterOperationSuccess(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("X-pass-after-succeed-operation", "true")

	abortAfterOperationSuccess := r.Header.Get("abort-after-operation")
	if abortAfterOperationSuccess == "true" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "abort-after-operation header is set to true"})
		return false
	}
	return true
}

func MiddlewareOnError(w http.ResponseWriter, r *http.Request, err error) bool {
	w.Header().Set("X-pass-on-error", "true")

	abortOnError := r.Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case assets.CustomError:
			operationErr = e.Message
		default:
			operationErr = err.Error()
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}

func MiddlewareOnError2(w http.ResponseWriter, r *http.Request, err error) bool {
	w.Header().Set("X-pass-on-error-2", "true")
	return true
}

func MiddlewareOnValidationError(w http.ResponseWriter, r *http.Request, err error) bool {
	w.Header().Set("X-pass-error-validation", "true")

	abortOnError := r.Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case error:
			operationErr = e.Error()
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}

func MiddlewareOnOutputValidationError(w http.ResponseWriter, r *http.Request, err error) bool {
	w.Header().Set("X-pass-output-validation", "true")

	returnNull := r.Header.Get("x-return-null")
	if returnNull == "true" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nil)
		return false
	}

	abortOnError := r.Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case error:
			operationErr = e.Error()
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}
