package handlers

import "github.com/merkulovlad/avito-internship-test/internal/user"

type UserHandler struct {
	service user.UserService
}

func NewUserHandler(service user.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}
