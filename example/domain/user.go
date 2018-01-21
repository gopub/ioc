package domain

import "github.com/natande/gox"

type User struct {
	ID   int
	Name string
}

type UserService struct {
	UserRepo UserRepo `inject:"auto"`
}

func (s *UserService) Init() {
	u, err := s.UserRepo.GetUser(1)
	if err != nil {
		gox.LogError(err)
	} else {
		gox.LogInfo(u.Name)
	}
}
