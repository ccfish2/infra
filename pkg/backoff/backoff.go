package backoff

import (
	"math"
	"time"

	"github.com/ccfish2/infra/pkg/rand"
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
