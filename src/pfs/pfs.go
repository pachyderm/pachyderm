package pfs

import "fmt"

var (
	ErrDiscoveryNetworkFailure     = NewError(ErrorCode_ERROR_CODE_DISCOVERY_NETWORK_FAILURE, "")
	ErrDiscoveryNotFound           = NewError(ErrorCode_ERROR_CODE_DISCOVERY_NOT_FOUND, "")
	ErrDiscoveryNotDirectory       = NewError(ErrorCode_ERROR_CODE_DISCOVERY_NOT_VALUE, "")
	ErrDiscoveryNotValue           = NewError(ErrorCode_ERROR_CODE_DISCOVERY_NOT_VALUE, "")
	ErrDiscoveryKeyAlreadyExists   = NewError(ErrorCode_ERROR_CODE_DISCOVERY_KEY_ALREADY_EXISTS, "")
	ErrDiscoveryPreconditionNotMet = NewError(ErrorCode_ERROR_CODE_PRECONDITION_NOT_MET, "")
)

func (e *Error) Error() string {
	return fmt.Sprintf("%+v", e)
}

func NewError(errorCode ErrorCode, format string, args ...interface{}) error {
	return &Error{
		ErrorCode: errorCode,
		Value:     fmt.Sprintf(format, args...),
	}
}

func Errorf(format string, args ...interface{}) error {
	return NewError(ErrorCode_ERROR_CODE_UNKNOWN, format, args...)
}

func GetErrorCode(err error) (ErrorCode, bool) {
	if typedErr, ok := err.(*Error); ok {
		return typedErr.ErrorCode, true
	}
	return -1, false
}

func VersionString(version *Version) string {
	return fmt.Sprintf("%d.%d.%d%s", version.Major, version.Minor, version.Micro, version.Additional)
}
