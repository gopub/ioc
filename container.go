package ioc

type Container interface {
	RegisterValue(name string, value interface{}) bool

	RegisterSingleton(prototype interface{}) string
	RegisterTransient(prototype interface{}) string

	RegisterTransientCreator(name string, creator Creator) bool
	RegisterSingletonCreator(name string, creator Creator) bool

	Contains(name string) bool
	RegisterAlias(name string, alias ...string) bool
	GetAlias(name string) []string

	Resolve(name string) interface{}

	//NewChildContainer() Container
}

var rootContainer = NewContainer()

func RootContainer() Container {
	return rootContainer
}

func NewContainer() Container {
	c := &containerImpl{}
	c.nameToRegistryIndex = make(map[string]int, 10)
	c.factory = NewFactory()
	return c
}

func RegisterValue(name string, value interface{}) bool {
	return rootContainer.RegisterValue(name, value)
}

func RegisterSingleton(prototype interface{}) string {
	return rootContainer.RegisterSingleton(prototype)
}

func RegisterTransient(prototype interface{}) string {
	return rootContainer.RegisterTransient(prototype)
}

func RegisterTransientCreator(name string, creator Creator) bool {
	return rootContainer.RegisterTransientCreator(name, creator)
}

func RegisterSingletonCreator(name string, creator Creator) bool {
	return rootContainer.RegisterSingletonCreator(name, creator)
}

func Contains(name string) bool {
	return rootContainer.Contains(name)
}

func RegisterAlias(name string, alias ...string) bool {
	return rootContainer.RegisterAlias(name, alias...)
}

func GetAlias(name string) []string {
	return rootContainer.GetAlias(name)
}

func Resolve(name string) interface{} {
	return rootContainer.Resolve(name)
}
