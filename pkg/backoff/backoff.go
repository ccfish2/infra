package backoff

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/ccfish2/infra/pkg/rand"
	"github.com/google/uuid"
)

var (
	randGen = rand.NewSafeRand(time.Now().UnixNano()) // thread-safe random number generator
)

type NodeManager interface {
	ClusterSizeDependantInterval(baseInterval time.Duration) time.Duration
}

// Backoff is an interface for implementing different backoff strategies.

type Exponential struct {
	Min              time.Duration
	Max              time.Duration
	Factor           float64
	Jitter           bool
	NodeManager      NodeManager
	Name             string
	ResetAfter       time.Duration
	lastBackoffStart time.Time
	attempt          int
	Logger           *slog.Logger
}

func CalculateDuration(min, max time.Duration, factor float64, jitter bool, failures int) time.Duration {
	minFloat := float64(min)
	maxFloat := float64(max)

	t := minFloat * math.Pow(factor, float64(failures))
	if max != time.Duration(0) && t > maxFloat {
		t = maxFloat
	}

	if jitter {
		t = randGen.Float64()*(t-minFloat) + minFloat
	}

	return time.Duration(t)
}

func (b *Exponential) Duration(attempt int) time.Duration {
	if b.Name == "" {
		b.Name = uuid.New().String()
	}
	min := time.Duration(1) * time.Second
	if b.Min != time.Duration(0) {
		min = b.Min
	}

	factor := float64(2)
	if b.Factor != float64(0) {
		factor = b.Factor
	}

	t := CalculateDuration(min, b.Max, factor, b.Jitter, attempt)
	if b.NodeManager != nil {
		t = b.NodeManager.ClusterSizeDependantInterval(t)
	}
	if b.Max != time.Duration(0) && t > b.Max {
		t = b.Max
	}
	return t
}

// Wait waits for the required time using an exponential backoff
func (b *Exponential) Wait(ctx context.Context) error {
	if resetDuration := b.ResetAfter; resetDuration != time.Duration(0) && resetDuration > b.Max {
		if !b.lastBackoffStart.IsZero() {
			if time.Since(b.lastBackoffStart) > resetDuration {
				b.Reset()
			}
		}
	}

	b.lastBackoffStart = time.Now()
	b.attempt++
	t := b.Duration(b.attempt)

	b.Logger.Debug("Sleeping with exponential backoff",
		logfields.Duration, t,
		logfields.Attempt, b.attempt,
		logfields.Name, b.Name,
	)

	select {
	case <-ctx.Done():
		return fmt.Errorf("exponential backoff cancelled via context: %w", ctx.Err())
	case <-time.After(t):
	}

	return nil
}

// Reset backoff attempt counter
func (b *Exponential) Reset() {
	b.attempt = 0
}

// Attempt returns the number of attempts since the last reset.
func (b *Exponential) Attempt() int {
	return b.attempt
}
