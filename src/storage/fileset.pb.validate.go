// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: storage/fileset.proto

package storage

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

// Validate checks the field values on AppendFile with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *AppendFile) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AppendFile with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in AppendFileMultiError, or
// nil if none found.
func (m *AppendFile) ValidateAll() error {
	return m.validate(true)
}

func (m *AppendFile) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Path

	if all {
		switch v := interface{}(m.GetData()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, AppendFileValidationError{
					field:  "Data",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, AppendFileValidationError{
					field:  "Data",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetData()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return AppendFileValidationError{
				field:  "Data",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return AppendFileMultiError(errors)
	}

	return nil
}

// AppendFileMultiError is an error wrapping multiple validation errors
// returned by AppendFile.ValidateAll() if the designated constraints aren't met.
type AppendFileMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m AppendFileMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m AppendFileMultiError) AllErrors() []error { return m }

// AppendFileValidationError is the validation error returned by
// AppendFile.Validate if the designated constraints aren't met.
type AppendFileValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AppendFileValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AppendFileValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AppendFileValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AppendFileValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AppendFileValidationError) ErrorName() string { return "AppendFileValidationError" }

// Error satisfies the builtin error interface
func (e AppendFileValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAppendFile.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AppendFileValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AppendFileValidationError{}

// Validate checks the field values on DeleteFile with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeleteFile) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteFile with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeleteFileMultiError, or
// nil if none found.
func (m *DeleteFile) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteFile) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Path

	if len(errors) > 0 {
		return DeleteFileMultiError(errors)
	}

	return nil
}

// DeleteFileMultiError is an error wrapping multiple validation errors
// returned by DeleteFile.ValidateAll() if the designated constraints aren't met.
type DeleteFileMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteFileMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteFileMultiError) AllErrors() []error { return m }

// DeleteFileValidationError is the validation error returned by
// DeleteFile.Validate if the designated constraints aren't met.
type DeleteFileValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteFileValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteFileValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteFileValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteFileValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteFileValidationError) ErrorName() string { return "DeleteFileValidationError" }

// Error satisfies the builtin error interface
func (e DeleteFileValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteFile.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteFileValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteFileValidationError{}

// Validate checks the field values on CreateFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CreateFilesetRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CreateFilesetRequestMultiError, or nil if none found.
func (m *CreateFilesetRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateFilesetRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	switch v := m.Modification.(type) {
	case *CreateFilesetRequest_AppendFile:
		if v == nil {
			err := CreateFilesetRequestValidationError{
				field:  "Modification",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetAppendFile()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, CreateFilesetRequestValidationError{
						field:  "AppendFile",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, CreateFilesetRequestValidationError{
						field:  "AppendFile",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetAppendFile()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return CreateFilesetRequestValidationError{
					field:  "AppendFile",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *CreateFilesetRequest_DeleteFile:
		if v == nil {
			err := CreateFilesetRequestValidationError{
				field:  "Modification",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetDeleteFile()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, CreateFilesetRequestValidationError{
						field:  "DeleteFile",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, CreateFilesetRequestValidationError{
						field:  "DeleteFile",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetDeleteFile()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return CreateFilesetRequestValidationError{
					field:  "DeleteFile",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return CreateFilesetRequestMultiError(errors)
	}

	return nil
}

// CreateFilesetRequestMultiError is an error wrapping multiple validation
// errors returned by CreateFilesetRequest.ValidateAll() if the designated
// constraints aren't met.
type CreateFilesetRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateFilesetRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateFilesetRequestMultiError) AllErrors() []error { return m }

// CreateFilesetRequestValidationError is the validation error returned by
// CreateFilesetRequest.Validate if the designated constraints aren't met.
type CreateFilesetRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateFilesetRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateFilesetRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateFilesetRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateFilesetRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateFilesetRequestValidationError) ErrorName() string {
	return "CreateFilesetRequestValidationError"
}

// Error satisfies the builtin error interface
func (e CreateFilesetRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateFilesetRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateFilesetRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateFilesetRequestValidationError{}

// Validate checks the field values on CreateFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CreateFilesetResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CreateFilesetResponseMultiError, or nil if none found.
func (m *CreateFilesetResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateFilesetResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FilesetId

	if len(errors) > 0 {
		return CreateFilesetResponseMultiError(errors)
	}

	return nil
}

// CreateFilesetResponseMultiError is an error wrapping multiple validation
// errors returned by CreateFilesetResponse.ValidateAll() if the designated
// constraints aren't met.
type CreateFilesetResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateFilesetResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateFilesetResponseMultiError) AllErrors() []error { return m }

// CreateFilesetResponseValidationError is the validation error returned by
// CreateFilesetResponse.Validate if the designated constraints aren't met.
type CreateFilesetResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateFilesetResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateFilesetResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateFilesetResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateFilesetResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateFilesetResponseValidationError) ErrorName() string {
	return "CreateFilesetResponseValidationError"
}

// Error satisfies the builtin error interface
func (e CreateFilesetResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateFilesetResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateFilesetResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateFilesetResponseValidationError{}

// Validate checks the field values on RenewFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *RenewFilesetRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RenewFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// RenewFilesetRequestMultiError, or nil if none found.
func (m *RenewFilesetRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *RenewFilesetRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FilesetId

	// no validation rules for TtlSeconds

	if len(errors) > 0 {
		return RenewFilesetRequestMultiError(errors)
	}

	return nil
}

// RenewFilesetRequestMultiError is an error wrapping multiple validation
// errors returned by RenewFilesetRequest.ValidateAll() if the designated
// constraints aren't met.
type RenewFilesetRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RenewFilesetRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RenewFilesetRequestMultiError) AllErrors() []error { return m }

// RenewFilesetRequestValidationError is the validation error returned by
// RenewFilesetRequest.Validate if the designated constraints aren't met.
type RenewFilesetRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RenewFilesetRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RenewFilesetRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RenewFilesetRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RenewFilesetRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RenewFilesetRequestValidationError) ErrorName() string {
	return "RenewFilesetRequestValidationError"
}

// Error satisfies the builtin error interface
func (e RenewFilesetRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRenewFilesetRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RenewFilesetRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RenewFilesetRequestValidationError{}

// Validate checks the field values on ComposeFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ComposeFilesetRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ComposeFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ComposeFilesetRequestMultiError, or nil if none found.
func (m *ComposeFilesetRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ComposeFilesetRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for TtlSeconds

	if len(errors) > 0 {
		return ComposeFilesetRequestMultiError(errors)
	}

	return nil
}

// ComposeFilesetRequestMultiError is an error wrapping multiple validation
// errors returned by ComposeFilesetRequest.ValidateAll() if the designated
// constraints aren't met.
type ComposeFilesetRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ComposeFilesetRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ComposeFilesetRequestMultiError) AllErrors() []error { return m }

// ComposeFilesetRequestValidationError is the validation error returned by
// ComposeFilesetRequest.Validate if the designated constraints aren't met.
type ComposeFilesetRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ComposeFilesetRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ComposeFilesetRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ComposeFilesetRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ComposeFilesetRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ComposeFilesetRequestValidationError) ErrorName() string {
	return "ComposeFilesetRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ComposeFilesetRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sComposeFilesetRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ComposeFilesetRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ComposeFilesetRequestValidationError{}

// Validate checks the field values on ComposeFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ComposeFilesetResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ComposeFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ComposeFilesetResponseMultiError, or nil if none found.
func (m *ComposeFilesetResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ComposeFilesetResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FilesetId

	if len(errors) > 0 {
		return ComposeFilesetResponseMultiError(errors)
	}

	return nil
}

// ComposeFilesetResponseMultiError is an error wrapping multiple validation
// errors returned by ComposeFilesetResponse.ValidateAll() if the designated
// constraints aren't met.
type ComposeFilesetResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ComposeFilesetResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ComposeFilesetResponseMultiError) AllErrors() []error { return m }

// ComposeFilesetResponseValidationError is the validation error returned by
// ComposeFilesetResponse.Validate if the designated constraints aren't met.
type ComposeFilesetResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ComposeFilesetResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ComposeFilesetResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ComposeFilesetResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ComposeFilesetResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ComposeFilesetResponseValidationError) ErrorName() string {
	return "ComposeFilesetResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ComposeFilesetResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sComposeFilesetResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ComposeFilesetResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ComposeFilesetResponseValidationError{}

// Validate checks the field values on ShardFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ShardFilesetRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ShardFilesetRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ShardFilesetRequestMultiError, or nil if none found.
func (m *ShardFilesetRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ShardFilesetRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FilesetId

	// no validation rules for NumFiles

	// no validation rules for SizeBytes

	if len(errors) > 0 {
		return ShardFilesetRequestMultiError(errors)
	}

	return nil
}

// ShardFilesetRequestMultiError is an error wrapping multiple validation
// errors returned by ShardFilesetRequest.ValidateAll() if the designated
// constraints aren't met.
type ShardFilesetRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ShardFilesetRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ShardFilesetRequestMultiError) AllErrors() []error { return m }

// ShardFilesetRequestValidationError is the validation error returned by
// ShardFilesetRequest.Validate if the designated constraints aren't met.
type ShardFilesetRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ShardFilesetRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ShardFilesetRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ShardFilesetRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ShardFilesetRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ShardFilesetRequestValidationError) ErrorName() string {
	return "ShardFilesetRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ShardFilesetRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sShardFilesetRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ShardFilesetRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ShardFilesetRequestValidationError{}

// Validate checks the field values on PathRange with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *PathRange) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on PathRange with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in PathRangeMultiError, or nil
// if none found.
func (m *PathRange) ValidateAll() error {
	return m.validate(true)
}

func (m *PathRange) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Lower

	// no validation rules for Upper

	if len(errors) > 0 {
		return PathRangeMultiError(errors)
	}

	return nil
}

// PathRangeMultiError is an error wrapping multiple validation errors returned
// by PathRange.ValidateAll() if the designated constraints aren't met.
type PathRangeMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PathRangeMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PathRangeMultiError) AllErrors() []error { return m }

// PathRangeValidationError is the validation error returned by
// PathRange.Validate if the designated constraints aren't met.
type PathRangeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PathRangeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PathRangeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PathRangeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PathRangeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PathRangeValidationError) ErrorName() string { return "PathRangeValidationError" }

// Error satisfies the builtin error interface
func (e PathRangeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPathRange.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PathRangeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PathRangeValidationError{}

// Validate checks the field values on ShardFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ShardFilesetResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ShardFilesetResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ShardFilesetResponseMultiError, or nil if none found.
func (m *ShardFilesetResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ShardFilesetResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetShards() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ShardFilesetResponseValidationError{
						field:  fmt.Sprintf("Shards[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ShardFilesetResponseValidationError{
						field:  fmt.Sprintf("Shards[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ShardFilesetResponseValidationError{
					field:  fmt.Sprintf("Shards[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ShardFilesetResponseMultiError(errors)
	}

	return nil
}

// ShardFilesetResponseMultiError is an error wrapping multiple validation
// errors returned by ShardFilesetResponse.ValidateAll() if the designated
// constraints aren't met.
type ShardFilesetResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ShardFilesetResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ShardFilesetResponseMultiError) AllErrors() []error { return m }

// ShardFilesetResponseValidationError is the validation error returned by
// ShardFilesetResponse.Validate if the designated constraints aren't met.
type ShardFilesetResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ShardFilesetResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ShardFilesetResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ShardFilesetResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ShardFilesetResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ShardFilesetResponseValidationError) ErrorName() string {
	return "ShardFilesetResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ShardFilesetResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sShardFilesetResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ShardFilesetResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ShardFilesetResponseValidationError{}
