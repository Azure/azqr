package models

import "github.com/rs/zerolog/log"

type (
	ScanParams struct {
		ManagementGroups       []string
		Subscriptions          []string
		ResourceGroups         []string
		OutputName             string
		Stages                 *StageConfigs
		Xlsx                   bool
		Mask                   bool
		Csv                    bool
		Json                   bool
		Debug                  bool
		ScannerKeys            []string
		Filters                *Filters
		EnabledInternalPlugins map[string]bool
		// Profiling options (only effective when built with 'debug' tag)
		CPUProfile   string
		MemProfile   string
		TraceProfile string
		Sarif        bool
		Version      string
	}

	ScanArgs struct {
		Subscriptions  []string `json:"subscriptions,omitempty"`
		ResourceGroups []string `json:"resourceGroups,omitempty"`
		Services       []string `json:"services,omitempty"`
		Stages         []string `json:"stages,omitempty"`
		StageParams    []string `json:"stageParams,omitempty"`
		Mask           *bool    `json:"mask,omitempty"`
		// Xlsx and Json are opt-in and default to false when nil, the inverse of Mask (which
		// defaults to true for safety). File export is disabled by default so a typical MCP scan
		// does not write an Excel/JSON file to disk unless the caller explicitly requests one.
		Xlsx *bool `json:"xlsx,omitempty"`
		Json *bool `json:"json,omitempty"`
	}

	PluginScanArgs struct {
		Subscriptions  []string       `json:"subscriptions,omitempty"`
		ResourceGroups []string       `json:"resourceGroups,omitempty"`
		Mask           *bool          `json:"mask,omitempty"`
		Options        map[string]any `json:"options,omitempty"`
		// Xlsx and Json are opt-in and default to false when nil; see ScanArgs for the rationale.
		Xlsx *bool `json:"xlsx,omitempty"`
		Json *bool `json:"json,omitempty"`
	}
)

func NewScanParamsWithDefaults(args ScanArgs) *ScanParams {
	stages := NewStageConfigsWithDefaults()
	if len(args.Stages) > 0 {
		stages.ConfigureStages(args.Stages)
	}

	if err := stages.ApplyStageParams(args.StageParams); err != nil {
		log.Fatal().Err(err).Msg("failed applying stage parameters")
	}

	scannerKeys := args.Services
	filters := LoadFilters("", scannerKeys)

	mask := true
	if args.Mask != nil {
		mask = *args.Mask
	}

	xlsx := false
	if args.Xlsx != nil {
		xlsx = *args.Xlsx
	}

	json := false
	if args.Json != nil {
		json = *args.Json
	}

	return &ScanParams{
		ManagementGroups: []string{},
		Subscriptions:    args.Subscriptions,
		ResourceGroups:   args.ResourceGroups,
		OutputName:       "",
		Stages:           stages,
		Mask:             mask,
		Xlsx:             xlsx,
		Csv:              false,
		Json:             json,
		Debug:            false,
		ScannerKeys:      args.Services,
		Filters:          filters,
	}
}

func NewScanParamsForPlugins(args PluginScanArgs) *ScanParams {
	stages := NewStageConfigs()
	filters := LoadFilters("", []string{})
	mask := true
	if args.Mask != nil {
		mask = *args.Mask
	}

	xlsx := false
	if args.Xlsx != nil {
		xlsx = *args.Xlsx
	}

	json := false
	if args.Json != nil {
		json = *args.Json
	}

	return &ScanParams{
		ManagementGroups: []string{},
		Subscriptions:    args.Subscriptions,
		ResourceGroups:   args.ResourceGroups,
		OutputName:       "",
		Stages:           stages,
		Mask:             mask,
		Xlsx:             xlsx,
		Csv:              false,
		Json:             json,
		Debug:            false,
		ScannerKeys:      []string{},
		Filters:          filters,
	}
}
