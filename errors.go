package rusender

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Classes of API errors for errors.Is. An [*Error] matches the class of its HTTP status:
//
//	if errors.Is(err, rusender.ErrNotFound) { ... }
var (
	// ErrValidation: 400 and 422, the request failed validation; fields are in Error.Details.
	ErrValidation = errors.New("rusender: validation failed")
	// ErrAuthentication: 401, the key is missing, invalid or revoked.
	ErrAuthentication = errors.New("rusender: authentication failed")
	// ErrPaymentRequired: 402, not enough balance or plan limit.
	ErrPaymentRequired = errors.New("rusender: payment required")
	// ErrPermission: 403, the key lacks the required scope.
	ErrPermission = errors.New("rusender: permission denied")
	// ErrNotFound: 404, the resource was not found or belongs to another account.
	ErrNotFound = errors.New("rusender: not found")
	// ErrConflict: 409, a state conflict, or the Idempotency-Key was reused with a different body.
	ErrConflict = errors.New("rusender: conflict")
	// ErrRateLimit: 429, rate limit exceeded; wait Error.RetryAfter before retrying.
	ErrRateLimit = errors.New("rusender: rate limit exceeded")
)

var (
	// ErrTimeout: no response within the configured timeout.
	ErrTimeout = errors.New("rusender: request timed out")
	// ErrUnexpectedResponse: a response outside the API contract, such as an HTML page of a proxy with
	// status 200 or a list without items. It is never returned as empty data.
	ErrUnexpectedResponse = errors.New("rusender: unexpected response from RuSender API")
	// ErrMissingAPIKey: no API key was passed and RUSENDER_API_KEY is not set.
	ErrMissingAPIKey = errors.New(
		"rusender: API key is missing: pass it to NewClient or set the RUSENDER_API_KEY environment variable")
)

// ErrorDetail describes one invalid field of a request.
type ErrorDetail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
	Value  any    `json:"value,omitempty"`
}

// Error is an API response with a non-2xx HTTP status.
type Error struct {
	// StatusCode is the HTTP status.
	StatusCode int
	// Code is the machine-readable code from error.code, such as CONTACT_NOT_FOUND.
	Code    string
	Message string
	Details []ErrorDetail
	// RequestID is the server-side request identifier: include it when you contact support.
	RequestID string
	// RetryAfter is the wait from the Retry-After header, zero when the header is absent.
	RetryAfter time.Duration
	Header     http.Header
}

func (e *Error) Error() string {
	text := fmt.Sprintf("rusender: %s (HTTP %d, %s)", e.Message, e.StatusCode, e.Code)
	if e.RequestID != "" {
		text += ", request " + e.RequestID
	}
	return text
}

// Is matches the error class of the status: ErrValidation, ErrNotFound and so on.
func (e *Error) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return target == ErrValidation
	case http.StatusUnauthorized:
		return target == ErrAuthentication
	case http.StatusPaymentRequired:
		return target == ErrPaymentRequired
	case http.StatusForbidden:
		return target == ErrPermission
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusConflict:
		return target == ErrConflict
	case http.StatusTooManyRequests:
		return target == ErrRateLimit
	}
	return false
}

// NetworkError means the request did not reach the API or no response was received: DNS, connection, TLS.
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return "rusender: request to RuSender API failed: " + e.Err.Error()
}

func (e *NetworkError) Unwrap() error { return e.Err }
