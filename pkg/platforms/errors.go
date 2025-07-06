package platforms

import (
	"fmt"
)

type ErrorCode string

const (
	ErrAuthentication        ErrorCode = "authentication_failed"
	ErrNotFound             ErrorCode = "not_found"
	ErrInvalidInput         ErrorCode = "invalid_input"
	ErrPlatformAPI          ErrorCode = "platform_api_error"
	ErrSyncConflict         ErrorCode = "sync_conflict"
	ErrPlatformNotSupported ErrorCode = "platform_not_supported"
	ErrInvalidConfig        ErrorCode = "invalid_config"
	ErrRateLimited          ErrorCode = "rate_limited"
	ErrPermissionDenied     ErrorCode = "permission_denied"
	ErrNetworkError         ErrorCode = "network_error"
)

type PlatformError struct {
	Code     ErrorCode `json:"code"`
	Message  string    `json:"message"`
	Platform string    `json:"platform,omitempty"`
	TaskID   string    `json:"task_id,omitempty"`
	Cause    error     `json:"-"`
}

func NewPlatformError(code ErrorCode, platform, taskID string, cause error) *PlatformError {
	return &PlatformError{
		Code:     code,
		Message:  getErrorMessage(code),
		Platform: platform,
		TaskID:   taskID,
		Cause:    cause,
	}
}

func (e *PlatformError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	
	if e.Platform != "" {
		msg += fmt.Sprintf(" (platform: %s)", e.Platform)
	}
	
	if e.TaskID != "" {
		msg += fmt.Sprintf(" (task: %s)", e.TaskID)
	}
	
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	
	return msg
}

func (e *PlatformError) Unwrap() error {
	return e.Cause
}

func (e *PlatformError) Is(target error) bool {
	if pe, ok := target.(*PlatformError); ok {
		return e.Code == pe.Code
	}
	return false
}

func getErrorMessage(code ErrorCode) string {
	switch code {
	case ErrAuthentication:
		return "Authentication failed"
	case ErrNotFound:
		return "Resource not found"
	case ErrInvalidInput:
		return "Invalid input provided"
	case ErrPlatformAPI:
		return "Platform API error"
	case ErrSyncConflict:
		return "Synchronization conflict"
	case ErrPlatformNotSupported:
		return "Platform not supported"
	case ErrInvalidConfig:
		return "Invalid configuration"
	case ErrRateLimited:
		return "Rate limit exceeded"
	case ErrPermissionDenied:
		return "Permission denied"
	case ErrNetworkError:
		return "Network error"
	default:
		return "Unknown error"
	}
}

func IsAuthenticationError(err error) bool {
	var pe *PlatformError
	return err != nil && (err == &PlatformError{Code: ErrAuthentication} || 
		(pe != nil && pe.Code == ErrAuthentication))
}

func IsNotFoundError(err error) bool {
	var pe *PlatformError
	return err != nil && (err == &PlatformError{Code: ErrNotFound} || 
		(pe != nil && pe.Code == ErrNotFound))
}

func IsRateLimitError(err error) bool {
	var pe *PlatformError
	return err != nil && (err == &PlatformError{Code: ErrRateLimited} || 
		(pe != nil && pe.Code == ErrRateLimited))
}