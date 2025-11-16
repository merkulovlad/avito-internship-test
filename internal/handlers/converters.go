package handlers

import (
	"github.com/merkulovlad/avito-internship-test/internal/api"
	"github.com/merkulovlad/avito-internship-test/internal/domain"
)

// UserToApi converts a domain User to an API User.
func UserToApi(u *domain.User) *api.User {
	return &api.User{
		UserId:   string(u.ID),
		Username: u.Username,
		TeamName: string(u.TeamName),
		IsActive: u.IsActive,
	}
}
