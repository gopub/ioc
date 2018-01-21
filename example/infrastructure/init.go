package infrastructure

import (
	"github.com/natande/go-ioc"
	"github.com/natande/go-ioc/example/domain"
)

func init() {
	ioc.RootContainer.RegisterSingletonInterface((*domain.UserRepo)(nil), &UserRepo{})
}
