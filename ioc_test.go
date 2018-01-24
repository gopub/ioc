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

type Greeting struct {
	Text string
}

func (f *Greeting) Init() {
	f.Text = "Hello"
}

func TestResolve(t *testing.T) {
	ioc.RegisterSingleton(&Greeting{})

	g := ioc.Resolve(ioc.NameOf(&Greeting{})).(*Greeting)
	t.Log(g.Text)
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
	ioc.RegisterAlias(name, ioc.NameOf((*PlusService)(nil)))
	c := ioc.Resolve(ioc.NameOf(&Calculator{})).(*Calculator)
	if c.PlusService.Plus(1, 2) != 3 {
		t.FailNow()
	}
}
