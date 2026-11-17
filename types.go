package rusender

import (
	"bytes"
	"encoding/json"
	"time"
)

// Ptr returns a pointer to v: a value for an optional request field.
func Ptr[T any](v T) *T { return &v }

// String returns a pointer to v: a value for an optional string field.
func String(v string) *string { return &v }

// Int returns a pointer to v: a value for an optional integer field.
func Int(v int) *int { return &v }

// Bool returns a pointer to v: a value for an optional boolean field.
func Bool(v bool) *bool { return &v }

// Time returns a pointer to v: a value for an optional date-time field.
func Time(v time.Time) *time.Time { return &v }

// Nullable is a request field that can be left out, set to a value, or set to null to clear the value
// stored by the API. The zero value is left out.
type Nullable[T any] struct {
	value T
	state nullableState
}

type nullableState uint8

const (
	nullableUnset nullableState = iota
	nullableNull
	nullableValue
)

// Value returns a Nullable set to v.
func Value[T any](v T) Nullable[T] { return Nullable[T]{value: v, state: nullableValue} }

// Null returns a Nullable that sends null: the API clears the stored value.
func Null[T any]() Nullable[T] { return Nullable[T]{state: nullableNull} }

// IsSet reports whether the field is sent: as a value or as null.
func (n Nullable[T]) IsSet() bool { return n.state != nullableUnset }

// IsNull reports whether the field is sent as null.
func (n Nullable[T]) IsNull() bool { return n.state == nullableNull }

// Get returns the value and whether the field holds one.
func (n Nullable[T]) Get() (T, bool) { return n.value, n.state == nullableValue }

// MarshalJSON writes the value or null. An unset field is left out by the parameters that hold it.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.state != nullableValue {
		return []byte("null"), nil
	}
	return json.Marshal(n.value)
}

// UnmarshalJSON reads a value or null.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*n = Null[T]()
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*n = Value(v)
	return nil
}

// sent returns nil for an unset field, so that `omitempty` leaves it out.
func (n Nullable[T]) sent() *Nullable[T] {
	if !n.IsSet() {
		return nil
	}
	return &n
}

// keepEmpty returns nil for a nil slice and a pointer otherwise: an empty, non-nil slice is sent as [],
// a nil one is left out by `omitempty`.
func keepEmpty[T any](s []T) *[]T {
	if s == nil {
		return nil
	}
	return &s
}

// Pagination of a page/limit list.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// Page is one page of a paginated list. Request the next page with Page + 1.
type Page[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// CursorPage is one page of a cursor-paginated list. Request the next page with Cursor set to NextCursor.
type CursorPage[T any] struct {
	Items      []T     `json:"items"`
	HasMore    bool    `json:"hasMore"`
	NextCursor *string `json:"nextCursor"`
}

// PageParams selects a page of a list.
type PageParams struct {
	// Page number, starting from 1.
	Page int `json:"page,omitempty"`
	// Page size, 1–100, 20 by default.
	Limit int `json:"limit,omitempty"`
}

// SortDirection is the sort order of a list.
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// Status is the on/off state of a list, webhook or sender verification.
type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

// DeleteResult is the result of a delete operation.
type DeleteResult struct {
	Deleted bool `json:"deleted"`
}

// ExportFormat is the file format of a contact export.
type ExportFormat string

const (
	ExportCSV  ExportFormat = "csv"
	ExportJSON ExportFormat = "json"
)
