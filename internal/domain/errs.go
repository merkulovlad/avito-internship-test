package domain

import "github.com/merkulovlad/avito-internship-test/internal/api"

// ServiceError represents a domain-level error with a specific code.
type ServiceError struct {
	// Code is the error code from the API specification.
	Code api.ErrorResponseErrorCode
	// Message is a human-readable error message.
	Message string
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new ServiceError with the given code and message.
func NewServiceError(code api.ErrorResponseErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// Predefined service errors.
var (
	// ErrTeamAlreadyExists is returned when attempting to create a team that already exists.
	ErrTeamAlreadyExists = NewServiceError(api.TEAMEXISTS, "team_name already exists")
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = NewServiceError(api.NOTFOUND, "resource not found")
	// ErrPrMerged is returned when attempting to modify a merged pull request.
	ErrPrMerged = NewServiceError(api.PRMERGED, "cannot reassign on merged PR")
	// ErrNotAssigned is returned when a reviewer is not assigned to a pull request.
	ErrNotAssigned = NewServiceError(api.NOTASSIGNED, "reviewer is not assigned to this PR")
	// ErrNoCandidate is returned when no suitable replacement reviewer is available.
	ErrNoCandidate = NewServiceError(api.NOCANDIDATE, "no active replacement candidate in team")
	// ErrPrAlreadyExists is returned when attempting to create a pull request with an existing ID.
	ErrPrAlreadyExists = NewServiceError(api.PREXISTS, "PR id already exists")
)
