package ioc_test

import (
	"github.com/natande/go-ioc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type PageController struct {
	Title string `inject:"page_title"`
}

func TestInjectValue(t *testing.T) {
	title := "This is a page"
	ioc.RegisterValue("page_title", title)
	ioc.RegisterSingleton(&PageController{})

	controller := ioc.Resolve(ioc.NameOf(&PageController{})).(*PageController)
	if controller.Title != title {
		t.FailNow()
	}
}
