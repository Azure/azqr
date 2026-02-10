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
		Stdout                 bool
		Debug                  bool
		ScannerKeys            []string
		Filters                *Filters
		EnabledInternalPlugins map[string]bool
		// Profiling options (only effective when built with 'debug' tag)
		CPUProfile   string
		MemProfile   string
		TraceProfile string
	}

	ScanArgs struct {
		Subscriptions  []string `json:"subscriptions,omitempty"`
		ResourceGroups []string `json:"resourceGroups,omitempty"`
		Services       []string `json:"services,omitempty"`
		Stages         []string `json:"stages,omitempty"`
		StageParams    []string `json:"stageParams,omitempty"`
		Mask           *bool    `json:"mask,omitempty"`
	}

	PluginScanArgs struct {
		Subscriptions  []string `json:"subscriptions,omitempty"`
		ResourceGroups []string `json:"resourceGroups,omitempty"`
		Mask           *bool    `json:"mask,omitempty"`
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

	return &ScanParams{
		ManagementGroups: []string{},
		Subscriptions:    args.Subscriptions,
		ResourceGroups:   args.ResourceGroups,
		OutputName:       "",
		Stages:           stages,
		Mask:             mask,
		Csv:              false,
		Json:             false,
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

	return &ScanParams{
		ManagementGroups: []string{},
		Subscriptions:    args.Subscriptions,
		ResourceGroups:   args.ResourceGroups,
		OutputName:       "",
		Stages:           stages,
		Mask:             mask,
		Csv:              false,
		Json:             false,
		Debug:            false,
		ScannerKeys:      []string{},
		Filters:          filters,
	}
}
