// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"testing"
	"time"
)

func TestCostTimeRangePreviousMonth(t *testing.T) {
	now := time.Date(2026, time.February, 3, 12, 0, 0, 0, time.UTC)
	start, end := costTimeRange(now, true)

	wantStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)

	if !start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", start, wantStart)
	}

	if !end.Equal(wantEnd) {
		t.Fatalf("end = %v, want %v", end, wantEnd)
	}
}

func TestCostTimeRangeDefault(t *testing.T) {
	now := time.Date(2026, time.February, 3, 12, 0, 0, 0, time.UTC)
	start, end := costTimeRange(now, false)

	wantStart := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", start, wantStart)
	}

	if !end.Equal(now) {
		t.Fatalf("end = %v, want %v", end, now)
	}
}
