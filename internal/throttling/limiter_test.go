// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package throttling

import (
	"context"
	"testing"
	"time"
)

func TestWaitARM(t *testing.T) {
	ctx := context.Background()

	// Test that WaitARM doesn't return an error
	err := WaitARM(ctx)
	if err != nil {
		t.Errorf("WaitARM() returned unexpected error: %v", err)
	}
}

func TestWaitGraph(t *testing.T) {
	ctx := context.Background()

	// Test that WaitGraph doesn't return an error
	err := WaitGraph(ctx)
	if err != nil {
		t.Errorf("WaitGraph() returned unexpected error: %v", err)
	}
}

func TestWaitARM_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := WaitARM(ctx)
	if err == nil {
		t.Error("WaitARM() should return error when context is cancelled")
	}
	if err != context.Canceled {
		t.Errorf("WaitARM() returned %v, want context.Canceled", err)
	}
}

func TestWaitGraph_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := WaitGraph(ctx)
	if err == nil {
		t.Error("WaitGraph() should return error when context is cancelled")
	}
	if err != context.Canceled {
		t.Errorf("WaitGraph() returned %v, want context.Canceled", err)
	}
}

func TestWaitARM_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	err := WaitARM(ctx)
	if err == nil {
		t.Error("WaitARM() should return error when context times out")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("WaitARM() returned %v, want context.DeadlineExceeded", err)
	}
}

func TestWaitGraph_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	err := WaitGraph(ctx)
	if err == nil {
		t.Error("WaitGraph() should return error when context times out")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("WaitGraph() returned %v, want context.DeadlineExceeded", err)
	}
}

func TestARMLimiter_NotNil(t *testing.T) {
	if ARMLimiter == nil {
		t.Error("ARMLimiter should not be nil")
	}
}

func TestGraphLimiter_NotNil(t *testing.T) {
	if GraphLimiter == nil {
		t.Error("GraphLimiter should not be nil")
	}
}
