package domain

type UserID string

type User struct {
	ID       UserID
	Username string
	TeamName TeamName
	IsActive bool
}
