package domain

import "github.com/natande/go-ioc"

func init() {
	ioc.RootContainer.RegisterSingleton(&UserService{})
}
