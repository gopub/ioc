package ioc

type Initializer interface {
	Init()
}

type BeforeInjector interface {
	BeforeInject()
}

type AfterInjector interface {
	AfterInject()
}
