package rate

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type APILimiterParameters struct {
	EstimatedProcessDuration time.Duration
	AutoAdjust               bool
	MeanOver                 int
	ParallelRequests         int
	MaxParallelRequests      int
	MinParallelRequests      int
	RateLimit                rate.Limit
	RateBurst                int
	MinWaitDuration          time.Duration
	MaxWaitDuration          time.Duration
	Log                      bool
	DelayedAdjustmentFactor  bool
	SkipInitial              int
	MaxAdjustmentFacotr      int
}

type APILmiter struct {
	Parameters APILimiterParameters
}

func NewAPILimiterFromConfig(config string) (*APILmiter, error) {
	p := &APILimiterParameters{}
	if err := p.mergeUserConfig(config); err != nil {
		return nil, err
	}
	return &APILmiter{Parameters: *p}, nil
}

func (p *APILimiterParameters) mergeUserConfig(config string) error {
	tokens := strings.Split(config, ",")
	for _, token := range tokens {
		// Process each token to set parameters
		if token == "" {

			continue
		}
		st := strings.SplitN(token, ":", 2)
		if len(st) < 2 {
			return fmt.Errorf("invalid config format: %s", token)
		}
		if err := p.mergeUserConfigKeyValue(st[0], st[1]); err != nil {
			return err
		}
	}
	return nil
}

func (p *APILimiterParameters) mergeUserConfigKeyValue(key, value string) error {
	// Implement the logic to set parameters based on key-value pairs
	return nil
}
