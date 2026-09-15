// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package throttling

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// CostRequestTimeout bounds one Cost Management page, including pacing and retries.
const CostRequestTimeout = 15 * time.Minute

const costQuerySuffix = "/providers/microsoft.costmanagement/query"

type costScopeLimiter struct {
	limiter  *rate.Limiter
	lastUsed time.Time
	active   int
}

func (p *ThrottlingPolicy) acquireCostScope(req *http.Request) *costScopeLimiter {
	path := strings.TrimRight(req.URL.Path, "/")
	key := strings.ToLower(req.URL.Host + path[:len(path)-len(costQuerySuffix)])
	p.costMu.Lock()
	defer p.costMu.Unlock()
	now := time.Now()
	if p.costScopes == nil {
		p.costScopes = make(map[string]*costScopeLimiter)
	}
	if now.Sub(p.lastCostCleanup) >= CostRequestTimeout {
		for key, entry := range p.costScopes {
			if entry.active == 0 && now.Sub(entry.lastUsed) >= CostRequestTimeout {
				delete(p.costScopes, key)
			}
		}
		p.lastCostCleanup = now
	}
	entry := p.costScopes[key]
	if entry == nil {
		entry = &costScopeLimiter{limiter: rate.NewLimiter(rate.Every(20*time.Second), 1)}
		p.costScopes[key] = entry
	}
	entry.active++
	return entry
}

func (p *ThrottlingPolicy) releaseCostScope(entry *costScopeLimiter) {
	p.costMu.Lock()
	defer p.costMu.Unlock()
	entry.active--
	entry.lastUsed = time.Now()
}

// IsCostManagementQuery identifies query requests, including continuation pages.
func IsCostManagementQuery(req *http.Request) bool {
	path := strings.TrimRight(req.URL.Path, "/")
	return len(path) >= len(costQuerySuffix) && strings.EqualFold(path[len(path)-len(costQuerySuffix):], costQuerySuffix)
}

func (p *ThrottlingPolicy) throttleCostRequest(req *policy.Request) (*http.Response, error) {
	limiter := p.CostLimiter
	if limiter == nil {
		limiter = costLimiter
	}
	ctx := req.Raw().Context()
	scope := p.acquireCostScope(req.Raw())
	defer p.releaseCostScope(scope)
	// Wait for the scope first so its backlog cannot reserve the global bucket.
	if err := scope.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("cost management scope throttling wait failed: %w", err)
	}
	if err := limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("cost management throttling wait failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// A token and the deadline timer can become ready together.
	if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
		return nil, context.DeadlineExceeded
	}
	if req.Raw().Header.Get("ClientType") == "" {
		req.Raw().Header.Set("ClientType", "azqr")
	}
	resp, err := req.Next()
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		normalizeCostRetryAfter(resp.Header, time.Now())
	}
	return resp, err
}

// Published guidance names "client"; retain "clienttype" as a compatibility alias.
var costQuotaDimensions = [...]string{"entity", "tenant", "client", "clienttype"}

func retryDuration(value string, unit time.Duration) (time.Duration, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || n < 0 || n > math.MaxInt64/int64(unit) {
		return 0, false
	}
	return time.Duration(n) * unit, true
}

func normalizeCostRetryAfter(headers http.Header, now time.Time) {
	var longest time.Duration
	selected := ""
	consider := func(header string, unit time.Duration, allowDate bool) {
		for _, value := range headers.Values(header) {
			delay, valid := retryDuration(value, unit)
			if !valid && allowDate {
				if date, err := http.ParseTime(value); err == nil {
					delay, valid = max(date.Sub(now), 0), true
				}
			}
			if !valid {
				log.Debug().Str("header", header).Msg("Ignoring invalid Cost Management retry delay")
				continue
			}
			if delay > longest {
				longest, selected = delay, header
			}
		}
	}
	consider("Retry-After", time.Second, true)
	consider("retry-after-ms", time.Millisecond, false)
	consider("x-ms-retry-after-ms", time.Millisecond, false)
	consider("x-ms-ratelimit-microsoft.consumption-retry-after", time.Second, false)
	consider("x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after", time.Second, false)
	for _, dimension := range costQuotaDimensions {
		consider("x-ms-ratelimit-microsoft.costmanagement-"+dimension+"-retry-after", time.Second, false)
		remaining := headers.Get("x-ms-ratelimit-remaining-microsoft.costmanagement-" + dimension + "-requests")
		if remaining != "" {
			log.Debug().Str("dimension", dimension).Str("remaining", remaining).
				Msg("Cost Management quota remaining")
		}
	}
	consumed := headers.Get("x-ms-ratelimit-microsoft.costmanagement-qpu-consumed")
	remaining := headers.Get("x-ms-ratelimit-microsoft.costmanagement-qpu-remaining")
	if consumed != "" || remaining != "" {
		log.Debug().Str("qpuConsumed", consumed).Str("qpuRemaining", remaining).
			Msg("Cost Management QPU usage")
	}
	if longest > 0 {
		// This is the SDK's highest-priority retry header. Round up, never retry early.
		milliseconds := (longest-1)/time.Millisecond + 1
		headers.Set("retry-after-ms", strconv.FormatInt(int64(milliseconds), 10))
		log.Debug().Str("header", selected).Dur("retryAfter", longest).
			Msg("Applying Cost Management retry delay")
	}
}

type costTimeoutPolicy struct {
	timeout time.Duration
}

// NewCostTimeoutPolicy bounds a Cost Management operation or network attempt.
// Install the operation policy before SDK retries, and the attempt policy after pacing.
func NewCostTimeoutPolicy(timeout time.Duration) policy.Policy {
	return &costTimeoutPolicy{timeout: timeout}
}

func (p *costTimeoutPolicy) Do(req *policy.Request) (*http.Response, error) {
	if p.timeout <= 0 || !IsCostManagementQuery(req.Raw()) {
		return req.Next()
	}
	ctx, cancel := context.WithTimeout(req.Raw().Context(), p.timeout)
	resp, err := req.Clone(ctx).Next()
	if err != nil || resp == nil || resp.Body == nil {
		cancel()
		return resp, err
	}
	resp.Body = &cancelBody{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil {
		b.cancel()
	}
	return n, err
}

func (b *cancelBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}
