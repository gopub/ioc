package infrastructure

import (
	"github.com/natande/go-ioc/example/domain"
	"fmt"
)

type UserRepo struct {
}

func (r *UserRepo) GetUser(id int) (*domain.User, error) {
	u := &domain.User{
		ID:   id,
		Name: fmt.Sprintf("Tom%d", id),
	}
	return u, nil
}
