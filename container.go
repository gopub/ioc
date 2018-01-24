package ioc

type Container interface {
	RegisterValue(name string, value interface{})

	RegisterSingleton(prototype interface{}) string
	RegisterTransient(prototype interface{}) string

	RegisterTransientCreator(name string, creator Creator)
	RegisterSingletonCreator(name string, creator Creator)

	Contains(name string) bool
	RegisterAlias(name string, alias ...string)
	GetAlias(name string) []string

	Resolve(name string) interface{}

	//NewChildContainer() Container
}

var RootContainer = NewContainer()

func NewContainer() Container {
	c := &containerImpl{}
	c.nameToRegistryIndex = make(map[string]int, 10)
	c.factory = NewFactory()
	return c
}
