package hive

type Shutdowner interface {
	Shutdown(...ShutdownOption)
}

type ShutdownOption interface {
	apply(*shutdownOptions)
}

type shutdownOptions struct {
	err error
}
