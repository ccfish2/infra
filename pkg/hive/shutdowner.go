package hive

// stop hive
type Shutdowner interface {
	Shutdown(...ShutdownOption)
}

type ShutdownOption interface {
	apply(*shutdownOptions)
}

type shutdownOptions struct {
	err error
}

func ShutdownWithError(err error) ShutdownOption {
	return optionFunc(func(opts *shutdownOptions) {
		opts.err = err
	})
}

type optionFunc func(*shutdownOptions)

// apply implements ShutdownOption.
func (fn optionFunc) apply(opts *shutdownOptions) {
	fn(opts)
}
