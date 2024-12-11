package hive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/spf13/pflag"
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

func (StringSliceConfig) Flags(flags *pflag.FlagSet) {
	flags.StringSlice("spaces-flag", nil, "split by spaces via pflag")
	flags.StringSlice("commas-flag", nil, "split by commas via pflag")
	flags.StringSlice("spaces-map", nil, "split by spaces via configmap")
	flags.StringSlice("spaces-map", nil, "split by commas via configmap")

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
	commas := "foo, bar, baz"
	expected := []string{"foo", "bar", "baz"}

	flags.Set("spaces-flag", spaces)
	flags.Set("commas-flag", commas)

	flags.Set("string-flag", "foo, bar, baz")

	flags.Set("mixed", "foo bar, baz")

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
	assert.Equal(t, cfg.StringFlag, "foo, bar, baz", "unexpected StringFlag")
}
