package cell

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"go.uber.org/dig"

	// myself
	"github.com/ccfish2/infra/pkg/cidr"
	"github.com/ccfish2/infra/pkg/utils"
)

type Flagger interface {
	Flags(*pflag.FlagSet)
}

// advanced: generic is added
func Config[Cfg Flagger](def Cfg) Cell {
	c := &config[Cfg]{
		defaultConfig: def,
		flags:         pflag.NewFlagSet("", pflag.ContinueOnError),
	}
	def.Flags(c.flags)
	return c
}

// advanced: generic
type config[Cfg Flagger] struct {
	defaultConfig Cfg
	flags         *pflag.FlagSet
}

type AllSettings map[string]any

type configParams[Cfg Flagger] struct {
	dig.In
	AllSettings AllSettings
	Override    func(*Cfg) `optional:"true"`
}

func decoderConfig(target any) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           target,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","), // string->[]string is split by comma
			fixupStringSliceHookFunc,                // []string of length 1 is split again by whitespace

			mapstructure.StringToTimeDurationHookFunc(),
			stringToCIDRHookFunc,
			stringToMapHookFunc,
		),
		ZeroFields:  true,
		ErrorUnset:  true,
		ErrorUnused: true,
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(
				strings.ReplaceAll(mapKey, "-", ""),
				fieldName)
		},
	}
}

func (c *config[Cfg]) provideConfig(p configParams[Cfg]) (Cfg, error) {
	settings := p.AllSettings
	target := c.defaultConfig
	decoder, err := mapstructure.NewDecoder(decoderConfig(&target))
	if err != nil {
		return target, fmt.Errorf("failed to create config decoder: %w", err)
	}

	input := make(map[string]any)

	c.flags.VisitAll(func(f *pflag.Flag) {
		if v, ok := settings[f.Name]; ok {
			input[f.Name] = v
		} else {
			err = fmt.Errorf("internal error: %s not found from settings", f.Name)
		}
	})
	if err != nil {
		return target, err
	}
	if err := decoder.Decode(input); err != nil {
		return target, fmt.Errorf("failed to unmarshal config struct %T: %w.\n"+
			"Hint: field 'FooBar' matches flag 'foo-bar', or use tag `mapstructure:\"flag-name\"` to match field with flag",
			target, err)
	}

	if p.Override != nil {
		p.Override(&target)
	}

	return target, nil
}

// Apply implements Cell.
func (c *config[Cfg]) Apply(cont container) error {
	// register the flags
	err := cont.Invoke(
		func(allFlags *pflag.FlagSet) {
			allFlags.AddFlagSet(c.flags)
		})
	if err != nil {
		return err
	}
	return cont.Provide(c.provideConfig, dig.Export(true))
}

func (c *config[Cfg]) Info(cont container) (info Info) {
	cont.Invoke(func(cfg Cfg) {
		info = &InfoStruct{cfg}
	})
	return
}

func fixupStringSliceHookFunc(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() != reflect.Slice || to.Kind() != reflect.Slice {
		return data, nil
	}

	if from.Elem().Kind() != reflect.String || to.Elem().Kind() != reflect.String {
		return data, nil
	}

	raw := data.([]string)
	if len(raw) == 1 {
		return strings.Fields(raw[0]), nil
	}
	return raw, nil
}

func stringToMapHookFunc(from reflect.Kind, to reflect.Kind, data interface{}) (interface{}, error) {
	if from != reflect.String || to != reflect.Map {
		return data, nil
	}
	return utils.ToStringMapStringE(data.(string))
}

func stringToCIDRHookFunc(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() != reflect.String {
		return data, nil
	}
	s := data.(string)
	if to != reflect.TypeOf((*cidr.CIDR)(nil)) {
		return data, nil
	}
	return cidr.ParseCIDR(s)
}
