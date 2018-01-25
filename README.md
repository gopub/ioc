1. Register and Resolve
```
type Greeting struct {
	Text string
}

//Register
ioc.RegisterSingleton(&Greeting{})

//Resolve
g := ioc.Resolve(ioc.NameOf(&Greeting{})).(*Greeting)
//...
```

2. Register for Interface
``` 
type Calculator struct {
	PlusService PlusService `inject:""`
}

//This is an interface
type PlusService interface {
	Plus(a, b int) int
}


//PlusServiceImpl implements PlusService
type PlusServiceImpl struct {
}

func (p *PlusServiceImpl) Plus(a, b int) int {
	return a + b
}

//Register transient type Calculator
ioc.RegisterTransient(&Calculator{})

//Register singleton type PlusServiceImpl
name := ioc.RegisterSingleton(&PlusServiceImpl{})

//Register PlusService as alias of PlusServiceImpl
ioc.RegisterAliases(name, ioc.NameOf((*PlusService)(nil)))

//PlusServiceImpl will be injected into Calculator.PlusService
c := ioc.Resolve(ioc.NameOf(&Calculator{})).(*Calculator)

c.PlusService.Plus(1, 2)
```