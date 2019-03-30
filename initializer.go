package ioc

type Initializer interface {
	Init()
}

type BeforeInjecter interface {
	BeforeInject()
}

type AfterInjecter interface {
	AfterInject()
}
