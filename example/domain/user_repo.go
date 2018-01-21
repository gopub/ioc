package domain

type UserRepo interface {
	GetUser(id int) (*User, error)
}
