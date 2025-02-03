package hive_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/cidr"
	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RunRollBack(t *testing.T) {

	var started, stopped int
	h := hive.New(
		cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
			lc.Append(cell.Hook{
				OnStart: func(hc cell.HookContext) error {
					started++
					return nil
				},
				OnStop: func(cell.HookContext) error {
					stopped++
					return nil
				},
			})
			lc.Append(cell.Hook{
				OnStart: func(hc cell.HookContext) error {
					started++
					<-hc.Done()
					return hc.Err()
				},
				OnStop: func(hc cell.HookContext) error {
					t.Fatal("unexpected call")
					return nil
				},
			})
		}),
	)
	h.SetTimeouts(time.Millisecond, time.Minute)

	err := h.Run()
	assert.ErrorIs(t, err, context.DeadlineExceeded, "unexpected call")

	assert.Equal(t, 2, started)
	assert.Equal(t, 1, stopped)
}

func TestShutdown(t *testing.T) {
	h := hive.New(
		cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
			lc.Append(cell.Hook{
				OnStart: func(cell.HookContext) error {
					shutdowner.Shutdown()
					return nil
				}})
		}),
	)
	assert.NoError(t, h.Run(), "expected Run() to succeed")

	h = hive.New(
		cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
			go shutdowner.Shutdown()
		}),
	)
	assert.NoError(t, h.Run(), "expected Run() to succeed though ")

	h = hive.New(
		cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
			shutdowner.Shutdown()
		}),
	)
	assert.NoError(t, h.Run(), "expected Run() to succeed")

	shutdownErr := errors.New("shutdown error")

	h = hive.New(
		cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
			lc.Append(cell.Hook{
				OnStart: func(cell.HookContext) error {
					shutdowner.Shutdown(hive.ShutdownWithError(shutdownErr))
					return nil
				},
			})
		}),
	)
	assert.ErrorIs(t, h.Run(), shutdownErr, "expected Run() to fail with shutdownErr")
}

var shutdownOnStartCell = cell.Invoke(func(lc cell.Lifecycle, shutdowner hive.Shutdowner) {
	lc.Append(cell.Hook{
		OnStart: func(hc cell.HookContext) error {
			shutdowner.Shutdown()
			return nil
		},
	})
})

type SomeObject struct {
	X int
}

type OtherObject struct {
	Y int
}

func TestDecorate(t *testing.T) {
	invokded := false

	testCell := cell.Decorate(
		func(o *SomeObject) *SomeObject { return &SomeObject{X: o.X + 1} },
		cell.Invoke(func(o *SomeObject) error {
			if o.X != 2 {
				return errors.New("X != 2")
			}
			invokded = true
			return nil
		}),
	)

	hive := hive.New(
		cell.Provide(func() *SomeObject { return &SomeObject{1} }),

		// here *SomeObject is not decorated
		cell.Invoke(func(o *SomeObject) error {
			if o.X != 1 {
				return errors.New("X != 1")
			}
			return nil
		}),
		testCell,
		shutdownOnStartCell,
	)

	assert.NoError(t, hive.Run(), "expected Run() to succeed")
	assert.True(t, invokded, "expected decorated invoke function to be called")
}

func TestProvidePrivate(t *testing.T) {
	invoked := true

	testCell := cell.Module(
		"test",
		"Test Module",
		cell.ProvidePrivate(func() *SomeObject { return &SomeObject{10} }),
		cell.Invoke(func(*SomeObject) { invoked = true }),
	)

	err := hive.New(
		testCell,
		shutdownOnStartCell,
	).Run()
	assert.NoError(t, err, "expected Start to succeed")

	if !invoked {
		t.Fatal("expected invoke to be called, but it was not")
	}

	h := hive.New(
		testCell,
		// this wont work which invoke the function within the testCell module
		cell.Invoke(func(*SomeObject) {}),
		shutdownOnStartCell,
	)
	err = h.Start(context.TODO())
	assert.ErrorContains(t, err, "*hive_test.SomeObject", "expected Start to fail to find *SomeObject")
}

func TestGroup(t *testing.T) {
	sum := 0

	testCell := cell.Group(
		cell.Provide(func() *SomeObject { return &SomeObject{10} }),
		cell.Provide(func() *OtherObject { return &OtherObject{5} }),
	)
	err := hive.New(
		testCell,
		cell.Invoke(func(a *SomeObject, b *OtherObject) { sum = a.X + b.Y }),
		shutdownOnStartCell,
	).Run()
	assert.NoError(t, err, "expected Run to succeed")
	assert.Equal(t, 15, sum)
}

func TestProvideHealthReporter(t *testing.T) {
	testCell := cell.Module(
		"test",
		"Test Module",
		cell.Invoke(func(s cell.Scope) {
			hr := cell.GetHealthReporter(s, "test")
			hr.OK("all good")
			hr.Stopped("stopped")
		}),
	)
	testCell2 := cell.Module(
		"test2",
		"Test Module 2",
		cell.Invoke(func(s cell.Scope) {
			hr := cell.GetHealthReporter(s, "test")
			hr.Degraded("degraded", nil)
		}),
	)

	unknown := cell.Module(
		"unknown",
		"Reports on status",
		cell.Invoke(func(s cell.Scope) {}),
	)

	var chp cell.Health

	h := hive.New(
		testCell,
		testCell2,
		unknown,
		cell.Invoke(func(lc cell.Lifecycle, _ hive.Shutdowner, hp cell.Health) {
			lc.Append(cell.Hook{
				OnStop: func(hc cell.HookContext) error {
					chp = hp
					return nil
				},
			})
		}),
		shutdownOnStartCell,
	)

	assert.NoError(t, h.Run(), "expected Run to succeed")
	s1, err := chp.Get(cell.FullModuleID{"test"})
	assert.NoError(t, err)

	s2, err := chp.Get(cell.FullModuleID{"test2"})
	assert.NoError(t, err)

	s3, err := chp.Get(cell.FullModuleID{"unknown"})
	assert.NoError(t, err)
	assert.Len(t, chp.All(), 3, "expected two health reports")
	assert.Equal(t, cell.StatusOK, s1.Level())

	assert.Equal(t, cell.StatusDegraded, s2.Level())
	assert.Equal(t, cell.StatusUnknown, s3.Level())
}

func TestModuleID(t *testing.T) {
	invoked := false
	inner := cell.Module(
		"inner",
		"inner module",
		cell.Invoke(func(id cell.ModuleID, fid cell.FullModuleID) error {
			invoked = true
			if id != "inner" {
				return fmt.Errorf("inner id mismatch, expected 'inner', got %q", id)
			}
			if fid.String() != "outer.inner" {
				return fmt.Errorf("outer id mismatch, expected 'outer.inner, got %q", fid)
			}
			return nil
		}))

	outer := cell.Module("outer", "outer module", inner)

	err := hive.New(outer, shutdownOnStartCell).Run()
	assert.NoError(t, err, "expected Run to succeed")
	assert.True(t, invoked, "expected invoke to be called, but it was not")
}

type StringSliceConfig struct {
	SpacesFlag, CommasFlag []string
	SpacesMap, CommasMap   []string
	Mixed                  []string
	StringFlag             string
}

func (s StringSliceConfig) Flags(flags *pflag.FlagSet) {
	flags.StringSlice("spaces-flag", nil, "split by spaces via pflag")
	flags.StringSlice("commas-flag", nil, "split by commas via pflag")
	flags.StringSlice("spaces-map", nil, "split by spaces via configmap")
	flags.StringSlice("commas-map", nil, "split by commas via configmap")

	flags.String("mixed", "", "mixed")
	flags.String("string-flag", "", "plain string untouched")
}

func TestHiveStringSlice(t *testing.T) {
	var cfg StringSliceConfig
	testCell := cell.Module(
		"test",
		"Test Module",
		cell.Config(StringSliceConfig{}),
		cell.Invoke(func(c StringSliceConfig) {
			cfg = c
		}),
	)

	hive := hive.New(testCell)

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	hive.RegisterFlags(flags)

	spaces := "foo bar baz"
	commas := "foo,bar,baz"
	expected := []string{"foo", "bar", "baz"}

	flags.Set("spaces-flag", spaces)
	flags.Set("commas-flag", commas)

	flags.Set("string-flag", "foo,bar,baz")

	flags.Set("mixed", "foo bar,baz")

	hive.Viper().MergeConfigMap(
		map[string]any{
			"spaces-map": spaces,
			"commas-map": commas,
		})

	err := hive.Start(context.TODO())
	require.NoError(t, err, "expected Start to succeed")
	err = hive.Stop(context.TODO())
	require.NoError(t, err, "expected Stop to succeed")

	assert.ElementsMatch(t, cfg.SpacesFlag, expected, "unexpected SpacesFlag")
	assert.ElementsMatch(t, cfg.SpacesMap, expected, "unexpected SpacesMap")
	assert.ElementsMatch(t, cfg.CommasFlag, expected, "unexpected CommasFlag")
	assert.ElementsMatch(t, cfg.CommasMap, expected, "unexpected CommasMap")
	assert.ElementsMatch(t, cfg.Mixed, []string{"foo bar", "baz"}, "unexpected Mixed")
	assert.Equal(t, cfg.StringFlag, "foo,bar,baz", "unexpected StringFlag")
}

type CIDRSliceConfig struct {
	Foo []*cidr.CIDR
}

func (CIDRSliceConfig) Flags(flags *pflag.FlagSet) {
	flags.StringSlice("foo", nil, "foo")
}

func TestHiveCIDRSlice(t *testing.T) {
	var cfg CIDRSliceConfig
	testCell := cell.Module(
		"test",
		"Test Module",
		cell.Config(CIDRSliceConfig{}),
		cell.Invoke(func(c CIDRSliceConfig) {
			cfg = c
		}),
	)
	hive := hive.New(testCell)

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	hive.RegisterFlags(flags)
	flags.Set("foo", "1.2.3.4/24,2001:db8::/64")

	err := hive.Start(context.TODO())
	require.NoError(t, err, "expected Start to succeed")
	err = hive.Stop(context.TODO())
	require.NoError(t, err, "expected Stop to succeed")

	require.Len(t, cfg.Foo, 2)
	require.Equal(t, cidr.MustParseCIDR("1.2.3.4/24"), cfg.Foo[0], "Config.Foo not set correctly")
	require.Equal(t, cidr.MustParseCIDR("2001:db8::/64"), cfg.Foo[1], "Config.Foo not set correctly")
}

type MapConfig struct {
	Foo map[string]string
}

func (mf MapConfig) Flags(flags *pflag.FlagSet) {
	flags.StringToString("foo", nil, "foo")
}
func TestHiveStringMapConfig(t *testing.T) {
	runnable := func(setter func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper), expected map[string]string) func(t *testing.T) {
		return func(t *testing.T) {
			defer os.Unsetenv("DOLPHIN_FOO")

			var cfg MapConfig
			testcell := cell.Module(
				"testcell",
				"Test Cell",
				cell.Config(MapConfig{}),
				cell.Invoke(func(c MapConfig) {
					cfg = c
				}),
			)

			hive := hive.New(testcell)

			flags := pflag.NewFlagSet("", pflag.ContinueOnError)
			hive.RegisterFlags(flags)

			setter(t, flags, hive.Viper())

			err := hive.Start(context.Background())
			require.NoError(t, err)

			err = hive.Stop(context.Background())
			require.NoError(t, err)

			require.Equal(t, expected, cfg.Foo, "config.foo not set correctly")
		}
	}

	t.Run("UNSET", runnable(func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper) {
	}, map[string]string{}))

	t.Run("flag-kv", runnable(func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper) {
		require.NoError(t, flag.Set("foo", "foo=bar, baz=qux"))
		require.NoError(t, flag.Set("foo", "freq=thud"))
	}, map[string]string{"foo": "bar", "baz": "qux", "fred": "thud"}))

	t.Run("env-kv", runnable(func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper) {
		require.NoError(t, os.Setenv("DOLPHIN_FOO", "foo=bar,baz=qux"))
	}, map[string]string{"foo": "bar", "baz": "qux"}))

	t.Run("env-json", runnable(func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper) {
		require.NoError(t, os.Setenv("DOLPHIN_FOO", `{"foo":"bar", "baz":"qux"}`))
	}, map[string]string{"foo": "bar", "baz": "qux"}))

	t.Run("config-yaml", runnable(func(t *testing.T, flags *pflag.FlagSet, vp *viper.Viper) {
		vp.SetConfigType("yaml")
		reader := strings.NewReader("foo:\n  foo: bar\n  baz: qux")
		require.NoError(t, vp.ReadConfig(reader), "Failed Reading config file")
	}, map[string]string{"foo": "bar", "baz": "qux"}))

	t.Run("config-json", runnable(func(t *testing.T, pflag *pflag.FlagSet, vp *viper.Viper) {
		vp.SetConfigType("json")
		reqder := strings.NewReader(`{"foo": {"foo":"bar","baz":"qux"}}`)
		require.NoError(t, vp.ReadConfig(reqder), "Failed reading config file")
	}, map[string]string{"foo": "bar", "baz": "qux"}))

	t.Run("cm-json", runnable(func(t *testing.T, pflag *pflag.FlagSet, vp *viper.Viper) {
		require.NoError(t, vp.MergeConfigMap(map[string]interface{}{"foo": `{"foo":"bar","baz":"qux"}`}))
	}, map[string]string{"foo": "bar", "baz": "qux"}))

	t.Run("UNSET", runnable(func(t *testing.T, pflag *pflag.FlagSet, viper *viper.Viper) {
		require.NoError(t, pflag.Set("foo", "foo=bar,barz=qux"))
		require.NoError(t, pflag.Set("foo", "fred=thud"))
	}, map[string]string{
		"foo":  "bar",
		"barz": "qux",
		"fred": "thud",
	}))

}

type BadConfig struct {
	Bar string
}

func (BadConfig) Flags(pflags *pflag.FlagSet) {
	pflags.String("foo", "foobar", "foo")
}

func TestHiveBadConfig(t *testing.T) {
	testCell := cell.Module(
		"test",
		"Test Module",
		cell.Config(BadConfig{}),
		cell.Invoke(func(c BadConfig) {}),
	)

	hive := hive.New(testCell)

	err := hive.Start(context.TODO())
	assert.ErrorContains(t, err, "has invalid keys: foo", "expected 'invalid keys' error")
	assert.ErrorContains(t, err, "has unset fields: Bar", "expected 'unset fields' error")
}

type Config struct {
	Foo string
	Bar int
}

// Foo string Bar int
func (Config) Flags(pflg *pflag.FlagSet) {
	pflg.String("foo", "hello world", "sets the greeting")
	pflg.Int("bar", 123, "bar")
}

// Flags function
// pass in foo string
// pass in bar int
func TestHiveGood(t *testing.T) {
	// define cfg
	var cfg Config
	// creaate a test Cell
	testCell := cell.Module(
		"test",
		"Test Module",
		cell.Config(Config{}),
		cell.Invoke(func(c Config) {
			cfg = c
		}),
	)

	// build a hive using the testCell
	h := hive.New(testCell)

	// create pflag using pflag.ContinueOnError
	// register the flag with the hive
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	h.RegisterFlags(flags)

	// Test two ways of setting it
	// configure the value for foo and bar
	flags.Set("foo", "test")
	h.Viper().Set("bar", 13)

	err := h.Start(context.TODO())
	assert.NoError(t, err, "expected Start to succeed")

	// execute hive.Start, hive.Stop
	// ensure foo and bar are
	err = h.Stop(context.TODO())
	assert.NoError(t, err, "expected stop to succeed")
	assert.Equal(t, "test", cfg.Foo, "Config.foo not set correctly")
	assert.Equal(t, 13, cfg.Bar, "Config.bar not set correctly")
}
