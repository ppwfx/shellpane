package errutil

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type StatusError interface {
	Status() string
	PublicError() string
}

const (
	StatusNotFound = "not found"

	StatusFailedDelete = "failed delete"

	StatusAlreadyExists = "already exists"

	StatusUnknown = "unknown"

	StatusUnauthorized = "unauthorized"

	StatusUnexpected = "unexpected"

	StatusFailedLogin = "login failed"

	StatusInvalid = "invalid"

	StatusDecoding = "failed decoding"

	StatusEncoding = "failed encoding"

	StatusExpiredToken = "expired token"

	StatusUnexpectedHTTPStatusCode = "unexpected status code"
)

var httpStatusCodeByStatus = map[string]int{
	StatusNotFound:                 http.StatusNotFound,
	StatusFailedDelete:             http.StatusInternalServerError,
	StatusAlreadyExists:            http.StatusConflict,
	StatusUnknown:                  http.StatusInternalServerError,
	StatusUnexpected:               http.StatusInternalServerError,
	StatusDecoding:                 http.StatusBadRequest,
	StatusFailedLogin:              0,
	StatusInvalid:                  http.StatusUnprocessableEntity,
	StatusExpiredToken:             http.StatusUnauthorized,
	StatusUnexpectedHTTPStatusCode: http.StatusBadGateway,
	StatusUnauthorized:             http.StatusForbidden,
}

func GetHTTPStatusCode(err StatusError) int {
	return httpStatusCodeByStatus[err.Status()]
}

func hasStatusError(err error) bool {
	var statusError StatusError
	return errors.As(err, &statusError)
}

func validateError(err error) {
	switch {
	case err == nil:
		panic("errutil failed to validate error: error is nil")
	case hasStatusError(err):
		panic("errutil failed to validate error: status error can't wrap status error")
	}
}

func Nil() error {
	return NilError{}
}

type NilError struct{}

func (e NilError) Error() string {
	return "nil"
}

func NotFound(err error, entity string, id interface{}) error {
	validateError(err)

	return NotFoundError{Err: err, Entity: entity, ID: id}
}

type NotFoundError struct {
	Err    error
	Entity string
	ID     interface{}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("entity=%v with ID=%v not found: %v", e.Entity, e.ID, e.Err.Error())
}

func (e NotFoundError) PublicError() string {
	return fmt.Sprintf("Failed to find %v with ID %v", e.Entity, e.ID)
}

func (e NotFoundError) Unwrap() error {
	return e.Err
}

func (e NotFoundError) Status() string {
	return StatusNotFound
}

func FailedDelete(err error, entity string, id interface{}) error {
	validateError(err)
	return FailedDeleteError{Err: err, Entity: entity, ID: id}
}

type FailedDeleteError struct {
	Err    error
	Entity string
	ID     interface{}
}

func (e FailedDeleteError) Error() string {
	return fmt.Sprintf("FailedDelete: Entity: %v ID: %v err: %v", e.Entity, e.ID, e.Err.Error())
}

func (e FailedDeleteError) PublicError() string {
	return fmt.Sprintf("Failed to delete %v with ID %v", e.Entity, e.ID)
}

func (e FailedDeleteError) Unwrap() error {
	return e.Err
}

func (e FailedDeleteError) Status() string {
	return StatusFailedDelete
}

func AlreadyExists(err error, entity string, id string) error {
	validateError(err)
	return AlreadyExistsError{Err: err, Entity: entity, ID: id}
}

type AlreadyExistsError struct {
	StatusError
	Err    error
	Entity string
	ID     string
}

func (e AlreadyExistsError) Error() string {
	return fmt.Sprintf("AlreadyExists: Entity: %v ID: %v err: %v", e.Entity, e.ID, e.Err.Error())
}

func (e AlreadyExistsError) PublicError() string {
	return fmt.Sprintf("%v with ID %v already exists", e.Entity, e.ID)
}

func (e AlreadyExistsError) Unwrap() error {
	return e.Err
}

func (e AlreadyExistsError) Status() string {
	return StatusAlreadyExists
}

func Unknown(err error) error {
	validateError(err)
	return UnknownError{Err: err}
}

type UnknownError struct {
	StatusError
	Err error
}

func (e UnknownError) Error() string {
	return fmt.Sprintf("Unknown err: %v", e.Err.Error())
}

func (e UnknownError) PublicError() string {
	return fmt.Sprintf("Failed for unknown reason")
}

func (e UnknownError) Unwrap() error {
	return e.Err
}

func (e UnknownError) Status() string {
	return StatusUnknown
}

func FailedLogin(err error, id string) error {
	validateError(err)
	return FailedLoginError{Err: err, ID: id}
}

type FailedLoginError struct {
	StatusError
	Err error
	ID  string
}

func (e FailedLoginError) Error() string {
	return fmt.Sprintf("FailedLogin: ID: %v err: %v", e.ID, e.Err.Error())
}

func (e FailedLoginError) PublicError() string {
	return fmt.Sprintf("Failed to login")
}

func (e FailedLoginError) Unwrap() error {
	return e.Err
}

func (e FailedLoginError) Status() string {
	return StatusFailedLogin
}

func Unexpected(err error) error {
	validateError(err)
	return UnexpectedError{Err: err}
}

type UnexpectedError struct {
	StatusError
	Err error
}

func (e UnexpectedError) Error() string {
	return fmt.Sprintf("Unexpected err: %v", e.Err.Error())
}

func (e UnexpectedError) PublicError() string {
	return fmt.Sprintf("Failed unexpectedly")
}

func (e UnexpectedError) Unwrap() error {
	return e.Err
}

func (e UnexpectedError) Status() string {
	return StatusUnexpected
}

func Invalid(err error) error {
	validateError(err)
	return InvalidError{Err: err}
}

type InvalidError struct {
	StatusError
	Err error
}

func (e InvalidError) Error() string {
	return fmt.Sprintf("Invalid err: %v", e.Err.Error())
}

func (e InvalidError) PublicError() string {
	return fmt.Sprintf("Failed as invalid")
}

func (e InvalidError) Unwrap() error {
	return e.Err
}

func (e InvalidError) Status() string {
	return StatusInvalid
}

func Decoding(err error) error {
	validateError(err)
	return DecodingError{Err: err}
}

type DecodingError struct {
	StatusError
	Err      error
	Encoding string
}

func (e DecodingError) Error() string {
	return fmt.Sprintf("Decoding Encoding: %v err: %v", e.Encoding, e.Err.Error())
}

func (e DecodingError) PublicError() string {
	return fmt.Sprintf("Failed to decode %v", e.Encoding)
}

func (e DecodingError) Unwrap() error {
	return e.Err
}

func (e DecodingError) Status() string {
	return StatusDecoding
}

func Encoding(err error) error {
	validateError(err)
	return EncodingError{Err: err}
}

type EncodingError struct {
	StatusError
	Err error
}

func (e EncodingError) Error() string {
	return fmt.Sprintf("Encoding err: %v", e.Err.Error())
}

func (e EncodingError) PublicError() string {
	return fmt.Sprintf("Failed to encode")
}

func (e EncodingError) Unwrap() error {
	return e.Err
}

func (e EncodingError) Status() string {
	return StatusEncoding
}

func ExpiredToken(tokenName string, expiration time.Time) error {
	return ExpiredTokenError{TokenName: tokenName, Expiration: expiration}
}

type ExpiredTokenError struct {
	StatusError
	TokenName  string
	Expiration time.Time
}

func (e ExpiredTokenError) Error() string {
	return fmt.Sprintf("ExpiredToken Name: %v Since: %v", e.TokenName, e.Expiration.Sub(time.Now()))
}

func (e ExpiredTokenError) PublicError() string {
	return fmt.Sprintf("Failed as token expired")
}

func (e ExpiredTokenError) Status() string {
	return StatusExpiredToken
}

func ExpectHTTPStatusCode(resp *http.Response, expected int) (err error) {
	if resp.StatusCode != expected {
		return UnexpectedHTTPStatusCode(expected, resp.StatusCode, resp.Status)
	}

	return nil
}

func UnexpectedHTTPStatusCode(expected int, actual int, httpStatus string) error {
	return UnexpectedHTTPStatusCodeError{Expected: expected, Actual: actual, HTTPStatus: httpStatus}
}

type UnexpectedHTTPStatusCodeError struct {
	StatusError
	Expected   int
	Actual     int
	HTTPStatus string
}

func (e UnexpectedHTTPStatusCodeError) Error() string {
	return fmt.Sprintf("UnexpectedHTTPStatusCode Expected: %v Actual: %v HTTPStatus: %v", e.Expected, e.Actual, e.HTTPStatus)
}

func (e UnexpectedHTTPStatusCodeError) PublicError() string {
	return fmt.Sprintf("Failed as token expired")
}

func (e UnexpectedHTTPStatusCodeError) Status() string {
	return StatusUnexpectedHTTPStatusCode
}

func Unauthorized(err error) error {
	//validateError(err)
	return UnauthorizedError{Err: err}
}

type UnauthorizedError struct {
	Err error
}

func (e UnauthorizedError) Error() string {
	return fmt.Sprintf("Unauthorized: err: %v", e.Err.Error())
}

func (e UnauthorizedError) PublicError() string {
	return fmt.Sprintf("Unauthorized")
}

func (e UnauthorizedError) Unwrap() error {
	return e.Err
}

func (e UnauthorizedError) Status() string {
	return StatusUnauthorized
}
