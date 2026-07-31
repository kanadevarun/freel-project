package svcerror

import "fmt"

type NasConfigErrorCode int

// Base platform error codes
const (
	ErrInternal                   NasConfigErrorCode = 1000
	ErrUnknown                    NasConfigErrorCode = 1001
	ErrInvalidArgument            NasConfigErrorCode = 1002
	ErrResourceNotFound           NasConfigErrorCode = 1003
	ErrDBDeadLock                 NasConfigErrorCode = 1004
	ErrServiceUnavailable         NasConfigErrorCode = 1005
	ErrInsufficientResourceAccess NasConfigErrorCode = 1006
	ErrUIErrorMessage             NasConfigErrorCode = 1007
	ErrAuthV3ServiceUnavailable   NasConfigErrorCode = 1008
	
	// Sample codes provided in user request
	ErrDeleteProxy                NasConfigErrorCode = 1060
	ErrTriggerDiscovery           NasConfigErrorCode = 1061
	ErrJobParams                  NasConfigErrorCode = 1062
)

// errorMap is a map of error constants to a string message
// Every constant MUST have an entry in this map
var errorMap = map[NasConfigErrorCode]string{
	ErrInternal:                   "Internal error",
	ErrUnknown:                    "Unknown error",
	ErrInvalidArgument:            "Invalid argument error",
	ErrResourceNotFound:           "Resource not found",
	ErrDBDeadLock:                 "DB deadlock detected",
	ErrServiceUnavailable:         "Service Unavailable",
	ErrInsufficientResourceAccess: "Admin do not have access to any of the requested resource",
	ErrUIErrorMessage:             "Internal Application Error",
	ErrAuthV3ServiceUnavailable:   "AuthV3 service Unvailable",

	ErrDeleteProxy:      "Delete proxy failed",
	ErrTriggerDiscovery: "Trigger discovery failed",
	ErrJobParams:        "Invalid job params",
}

// ServiceError represents a standardized error format
type ServiceError struct {
	Code    NasConfigErrorCode
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%d: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// NewServiceError creates a new ServiceError based on the code
func NewServiceError(code NasConfigErrorCode) *ServiceError {
	msg, ok := errorMap[code]
	if !ok {
		msg = errorMap[ErrUnknown]
	}
	return &ServiceError{
		Code:    code,
		Message: msg,
	}
}

// WrapServiceError creates a new ServiceError wrapping an underlying error
func WrapServiceError(code NasConfigErrorCode, err error) *ServiceError {
	msg, ok := errorMap[code]
	if !ok {
		msg = errorMap[ErrUnknown]
	}
	return &ServiceError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}
