package ioc_test

import (
	"github.com/gopub/ioc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type PageController struct {
	Title string `inject:"page_title"`
}

type Shape interface {
	Area() float64
}

type Rectangle struct {
	w float64
	h float64
}

func (r *Rectangle) Area() float64 {
	return r.w * r.h
}

func TestNameOf(t *testing.T) {
	t.Log(ioc.NameOf(&Rectangle{}))
	t.Log(ioc.NameOf(Rectangle{}))
	t.Log(ioc.NameOf((*Shape)(nil)))
}

func TestResolveByName(t *testing.T) {
	ioc.RegisterSingleton(&Rectangle{})

	g := ioc.ResolveByName(ioc.NameOf(&Rectangle{})).(*Rectangle)
	t.Log(g.Area())
}

func TestInjectValue(t *testing.T) {
	title := "This is a page"
	ioc.RegisterValue("page_title", title)
	ioc.RegisterSingleton(&PageController{})

	controller := ioc.ResolveByName(ioc.NameOf(&PageController{})).(*PageController)
	if controller.Title != title {
		t.FailNow()
	}
}

type Calculator struct {
	PlusService PlusService `inject:""`
}

type PlusService interface {
	Plus(a, b int) int
}

type PlusServiceImpl struct {
}

func (p *PlusServiceImpl) Plus(a, b int) int {
	return a + b
}

func TestInjectInterface(t *testing.T) {
	ioc.RegisterTransient(&Calculator{})

	name := ioc.RegisterSingleton(&PlusServiceImpl{})
	ioc.RegisterAliases(name, ioc.NameOf((*PlusService)(nil)))
	c := ioc.ResolveByName(ioc.NameOf(&Calculator{})).(*Calculator)
	if c.PlusService.Plus(1, 2) != 3 {
		t.FailNow()
	}
}
