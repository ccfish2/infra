package option

import "sync"

type VerifyFunc func(key, value string) error
type ParseFunc func(value string) (OptionSetting, error)
type FormatFunc func(value OptionSetting) string

// option semantics
type Option struct {
	Define      string
	Description string
	Immutable   bool
	Requires    []string
	Parse       ParseFunc
	Format      FormatFunc
	Verify      VerifyFunc
}
type OptionSetting int

const (
	OptionDisabled OptionSetting = iota
	OptionEnabled
)

type OptionMap map[string]OptionSetting
type OptionLibrary map[string]*Option

type IntOptions struct {
	optsMU  sync.RWMutex
	Opts    OptionMap      `json:"map"`
	Library *OptionLibrary `json:"-"`
}

func NewIntOptions(lib *OptionLibrary) *IntOptions {
	return &IntOptions{
		Opts:    OptionMap{},
		Library: lib,
	}
}

// following is runtime options
const (
	PolicyTracing = "PolicyTracing"
)

// following is daemon options
var (
	specPolicyTracing = Option{
		Description: "Enable tracing when resolving policy (Debug)",
	}
	DaemonOptionLibrary = OptionLibrary{
		PolicyTracing: &specPolicyTracing,
	}
)
