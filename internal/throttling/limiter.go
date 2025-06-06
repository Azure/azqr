package throttling

import "time"

type Limiter struct {
	bucketCapacity int
	refillRate     int
	refillEvery    time.Duration
	refillWait     time.Duration
}

func NewLimiter(bucketCapacity, refillRate int, refillEvery, refillWait time.Duration) *Limiter {
	return &Limiter{
		bucketCapacity: bucketCapacity,
		refillRate:     refillRate,
		refillEvery:    refillEvery,
		refillWait:     refillWait,
	}
}

func (l *Limiter) Start() chan struct{} {
	bucketCapacity := l.bucketCapacity
	refillRate := l.refillRate

	// Create a channel to act as a burst limiter.
	// This will allow up to bucketCapacity requests at once.

	// Staggering queries to avoid throttling.
	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling#regional-throttling-and-token-bucket-algorithm
	// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
	burstLimiter := make(chan struct{}, bucketCapacity)

	// Fill the burstLimiter channel with initial tokens.
	for i := 0; i < bucketCapacity; i++ {
		burstLimiter <- struct{}{}
	}

	// Create a ticker to limit the rate of requests
	limiter := time.NewTicker(l.refillEvery)

	// Start a goroutine to send ticks to the burstLimiter channel
	go func() {
		for range limiter.C {
			for i := 0; i < refillRate; i++ {
				// Only add a token if the channel is not full to avoid blocking
				select {
				case burstLimiter <- struct{}{}:
					time.Sleep(l.refillWait) // Wait for the specified refill wait time
				default:
					// Channel is full, skip adding more tokens
				}
			}
		}
	}()
	return burstLimiter
}
