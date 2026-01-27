// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build debug

package profiling

import (
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/rs/zerolog/log"
)

// Profiler manages CPU, memory, and execution trace profiling
type Profiler struct {
	cpuFile   *os.File
	memFile   string
	traceFile *os.File
}

// NewProfiler creates a new profiler instance
func NewProfiler() *Profiler {
	return &Profiler{}
}

// StartCPUProfile starts CPU profiling to the specified file
func (p *Profiler) StartCPUProfile(filename string) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return err
	}

	p.cpuFile = f
	log.Info().Msgf("CPU profiling enabled, writing to: %s", filename)
	return nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile() {
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		if err := p.cpuFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close CPU profile file")
		}
		p.cpuFile = nil
	}
}

// SetMemoryProfile sets the memory profile filename (written at end)
func (p *Profiler) SetMemoryProfile(filename string) {
	p.memFile = filename
}

// WriteMemoryProfile writes the memory profile
func (p *Profiler) WriteMemoryProfile() error {
	if p.memFile == "" {
		return nil
	}

	f, err := os.Create(p.memFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	log.Info().Msgf("Memory profile written to: %s", p.memFile)
	return nil
}

// StartTrace starts execution tracing to the specified file
func (p *Profiler) StartTrace(filename string) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := trace.Start(f); err != nil {
		f.Close()
		return err
	}

	p.traceFile = f
	log.Info().Msgf("Execution trace enabled, writing to: %s", filename)
	return nil
}

// StopTrace stops execution tracing
func (p *Profiler) StopTrace() {
	if p.traceFile != nil {
		trace.Stop()
		if err := p.traceFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close trace profile file")
		}
		p.traceFile = nil
	}
}

// Cleanup ensures all profiling resources are properly closed
func (p *Profiler) Cleanup() {
	p.StopCPUProfile()
	p.StopTrace()
	if err := p.WriteMemoryProfile(); err != nil {
		log.Error().Err(err).Msg("Failed to write memory profile")
	}
}

// IsProfilingAvailable returns true if profiling is compiled in
func IsProfilingAvailable() bool {
	return true
}
