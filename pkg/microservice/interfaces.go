package microservice

type Starter interface {
	Microservice

	Start() error
}

type Microservice interface {
	Valid() error
	Setup() error
	Defer()
	PanicHandler()
}
