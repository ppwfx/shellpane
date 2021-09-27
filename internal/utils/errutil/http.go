package errutil

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

// HandlerFuncJSON creates a http.HandlerFunc
func HandlerFuncJSON(f func(w http.ResponseWriter, r *http.Request) (rsp interface{}, err error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rsp, err := f(w, r)
		HandleJSONResponse(w, r, rsp, err)
	}
}

// HandleJSONResponse handles the response and returns json
func HandleJSONResponse(w http.ResponseWriter, r *http.Request, rsp interface{}, err error) {
	logger := logutil.MustLoggerValue(r.Context())
	{
		err := r.Body.Close()
		if err != nil {
			err = errors.Wrap(err, "failed to close request body")
			if logger != nil {
				logger.Error(err)
			} else {
				log.Printf(err.Error())
			}
		}
	}

	status := http.StatusOK
	if err != nil {
		// determine status code
		customErrorStatusCode, isCustomErrorStatusCode := rsp.(CustomErrorStatusCode)
		var statusErr StatusError
		switch {
		case isCustomErrorStatusCode && customErrorStatusCode.GetErrorStatusCode() != 0:
			status = customErrorStatusCode.GetErrorStatusCode()
			break
		case errors.As(err, &statusErr):
			status = GetHTTPStatusCode(statusErr)
			break
		default:
			status = http.StatusInternalServerError
		}

		// determine public error
		var publicError string
		customPublicError, isCustomPublicError := rsp.(CustomPublicError)
		switch {
		case isCustomPublicError:
			publicError = customPublicError.GetPublicError()
			break
		case errors.As(err, &statusErr):
			publicError = statusErr.PublicError()
			break
		default:
			publicError = "Internal Server Error"
		}

		// determine error response
		errorRsp, isErrorRsp := rsp.(ErrorResponse)
		_, isCustomErrorRsp := rsp.(CustomErrorResponse)
		switch {
		case isCustomErrorRsp:
			break
		case isErrorRsp:
			errorRsp.SetResponseError(ResponseError{
				Message: publicError,
			})
			break
		default:
			rsp = &Response{
				Error: ResponseError{
					Message: publicError,
				},
			}
		}

		if logger != nil {
			logger.Error(err)
		} else {
			log.Printf(err.Error())
		}
	}

	successRsp, ok := rsp.(CustomSuccessStatusCode)
	if ok {
		status = successRsp.GetSuccessStatusCode()
	}

	w.WriteHeader(status)

	err = json.NewEncoder(w).Encode(rsp)
	if err != nil {
		err = errors.Wrap(err, "failed to json encode response")
		if logger != nil {
			logger.Error(err)
		} else {
			log.Printf(err.Error())
		}
	}
}

// ErrorResponse defines an interface
type ErrorResponse interface {
	SetResponseError(err ResponseError)
}

// ResponseError defines a ResponseError
type ResponseError struct {
	Code    string
	Message string
}

// Response defines a Response
type Response struct {
	Error ResponseError
}

// SetResponseError sets a ResponseError
func (r *Response) SetResponseError(err ResponseError) {
	r.Error = err
}

// CustomSuccessStatusCode defines an interface
type CustomSuccessStatusCode interface {
	GetSuccessStatusCode() int
}

// CustomErrorResponse defines an interface
type CustomErrorResponse interface {
	IsCustomErrorResponse() int
}

// CustomErrorStatusCode defines an interface
type CustomErrorStatusCode interface {
	GetErrorStatusCode() int
}

// CustomPublicError defines an interface
type CustomPublicError interface {
	GetPublicError() string
}

// CustomErrorData defines custom error data
type CustomErrorData struct {
	StatusCode  int
	PublicError string
}

// GetPublicError return a custom public error message
func (r CustomErrorData) GetPublicError() string {
	return r.PublicError
}

// GetErrorStatusCode returns a custom status code
func (r CustomErrorData) GetErrorStatusCode() int {
	return r.StatusCode
}

// CustomSuccessData defines custom success data
type CustomSuccessData struct {
	StatusCode int `json:",omitempty"`
}

// GetSuccessStatusCode returns a custom status code
func (r CustomSuccessData) GetSuccessStatusCode() int {
	return r.StatusCode
}
