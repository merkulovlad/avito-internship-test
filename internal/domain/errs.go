package domain

import "github.com/merkulovlad/avito-internship-test/internal/api"

type ServiceError struct {
	Code    api.ErrorResponseErrorCode
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(code api.ErrorResponseErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamAlreadyExists = NewServiceError(api.TEAMEXISTS, "team_name already exists")
	ErrNotFound          = NewServiceError(api.NOTFOUND, "resource not found")
	ErrPrMerged          = NewServiceError(api.PRMERGED, "cannot reassign on merged PR")
	ErrNotAssigned       = NewServiceError(api.NOTASSIGNED, "reviewer is not assigned to this PR")
	ErrNoCandidate       = NewServiceError(api.NOCANDIDATE, "no active replacement candidate in team")
	ErrPrAlreadyExists   = NewServiceError(api.PREXISTS, "PR id already exists")
)
