package gena

import "time"

type retryParams struct {
	maxRetries   int
	initialDelay time.Duration
	delayFactor  float64
	maxDelay     time.Duration
}

func (r retryParams) GetMaxRetries() int {
	if r.maxRetries == 0 {
		return 5
	}
	return r.maxRetries
}

func (r retryParams) GetInitialDelay() time.Duration {
	if r.initialDelay == 0 {
		return 1 * time.Second
	}
	return r.initialDelay
}

func (r retryParams) GetDelayFactor() float64 {
	if r.delayFactor == 0 {
		return 1.5
	}
	return r.delayFactor
}

func (r retryParams) GetMaxDelay() time.Duration {
	if r.maxDelay == 0 {
		return 60 * time.Second
	}
	return r.maxDelay
}
