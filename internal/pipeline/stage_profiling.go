// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/profiling"
	"github.com/rs/zerolog/log"
)

// ProfilingStage manages profiling for the scan execution
type ProfilingStage struct {
	*BaseStage
	profiler *profiling.Profiler
}

// NewProfilingStage creates a new profiling stage
func NewProfilingStage() *ProfilingStage {
	return &ProfilingStage{
		BaseStage: NewBaseStage("Profiling Setup", true),
		profiler:  profiling.NewProfiler(),
	}
}

func (s *ProfilingStage) Execute(ctx *ScanContext) error {
	// Check if profiling parameters are set
	if ctx.Params.CPUProfile != "" || ctx.Params.MemProfile != "" || ctx.Params.TraceProfile != "" {
		if !profiling.IsProfilingAvailable() {
			log.Warn().Msg("Profiling requested but not available. Build with 'debug' tag to enable profiling.")
			return nil
		}

		log.Info().Msg("Profiling enabled")

		// Start CPU profiling
		if err := s.profiler.StartCPUProfile(ctx.Params.CPUProfile); err != nil {
			log.Error().Err(err).Msg("Failed to start CPU profiling")
		}

		// Set memory profile destination
		s.profiler.SetMemoryProfile(ctx.Params.MemProfile)

		// Start execution trace
		if err := s.profiler.StartTrace(ctx.Params.TraceProfile); err != nil {
			log.Error().Err(err).Msg("Failed to start execution trace")
		}

		// Store profiler in context for cleanup
		ctx.Profiler = s.profiler
	}

	return nil
}

// ProfilingCleanupStage handles profiling cleanup after scan completion
type ProfilingCleanupStage struct {
	*BaseStage
}

// NewProfilingCleanupStage creates a new profiling cleanup stage
func NewProfilingCleanupStage() *ProfilingCleanupStage {
	return &ProfilingCleanupStage{
		BaseStage: NewBaseStage("Profiling Cleanup", true),
	}
}

func (s *ProfilingCleanupStage) Execute(ctx *ScanContext) error {
	if ctx.Profiler != nil {
		log.Debug().Msg("Cleaning up profiling resources")
		ctx.Profiler.Cleanup()
	}
	return nil
}
