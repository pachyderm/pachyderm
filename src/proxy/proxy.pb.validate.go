// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: proxy/proxy.proto

package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on ListenRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListenRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListenRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListenRequestMultiError, or
// nil if none found.
func (m *ListenRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ListenRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetChannel()) < 1 {
		err := ListenRequestValidationError{
			field:  "Channel",
			reason: "value length must be at least 1 bytes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ListenRequestMultiError(errors)
	}

	return nil
}

// ListenRequestMultiError is an error wrapping multiple validation errors
// returned by ListenRequest.ValidateAll() if the designated constraints
// aren't met.
type ListenRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListenRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListenRequestMultiError) AllErrors() []error { return m }

// ListenRequestValidationError is the validation error returned by
// ListenRequest.Validate if the designated constraints aren't met.
type ListenRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListenRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListenRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListenRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListenRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListenRequestValidationError) ErrorName() string { return "ListenRequestValidationError" }

// Error satisfies the builtin error interface
func (e ListenRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListenRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListenRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListenRequestValidationError{}

// Validate checks the field values on ListenResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListenResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListenResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListenResponseMultiError,
// or nil if none found.
func (m *ListenResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ListenResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Extra

	if len(errors) > 0 {
		return ListenResponseMultiError(errors)
	}

	return nil
}

// ListenResponseMultiError is an error wrapping multiple validation errors
// returned by ListenResponse.ValidateAll() if the designated constraints
// aren't met.
type ListenResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListenResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListenResponseMultiError) AllErrors() []error { return m }

// ListenResponseValidationError is the validation error returned by
// ListenResponse.Validate if the designated constraints aren't met.
type ListenResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListenResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListenResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListenResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListenResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListenResponseValidationError) ErrorName() string { return "ListenResponseValidationError" }

// Error satisfies the builtin error interface
func (e ListenResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListenResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListenResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListenResponseValidationError{}
