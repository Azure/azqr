// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build !debug

package profiling

// Profiler is a no-op implementation when profiling is disabled
type Profiler struct{}

// NewProfiler creates a no-op profiler
func NewProfiler() *Profiler {
	return &Profiler{}
}

// StartCPUProfile is a no-op
func (p *Profiler) StartCPUProfile(filename string) error {
	return nil
}

// StopCPUProfile is a no-op
func (p *Profiler) StopCPUProfile() {}

// SetMemoryProfile is a no-op
func (p *Profiler) SetMemoryProfile(filename string) {}

// WriteMemoryProfile is a no-op
func (p *Profiler) WriteMemoryProfile() error {
	return nil
}

// StartTrace is a no-op
func (p *Profiler) StartTrace(filename string) error {
	return nil
}

// StopTrace is a no-op
func (p *Profiler) StopTrace() {}

// Cleanup is a no-op
func (p *Profiler) Cleanup() {}

// IsProfilingAvailable returns false when profiling is not compiled in
func IsProfilingAvailable() bool {
	return false
}
