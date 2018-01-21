package main

import (
	"fmt"
	"github.com/natande/go-ioc"
	"github.com/natande/go-ioc/example/domain"
	_ "github.com/natande/go-ioc/example/infrastructure"
)

func main() {
	ioc.RootContainer.Build()
	userService := ioc.RootContainer.GetObject(ioc.NameOfObject(&domain.UserService{})).(*domain.UserService)
	fmt.Println(userService)
}
