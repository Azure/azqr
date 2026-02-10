package models

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
		Stages         []string `json:"stages,omitempty"       jsonschema:"Optional scan stages. Default-on: graph, diagnostics, advisor, defender. Default-off (must be named to enable): policy, defender-recommendations, cost, arc. Use bare name or '+' prefix to enable (e.g. 'policy', '+defender-recommendations'). Use '-' prefix to disable a default-on stage (e.g. '-advisor')."`
		StageParams    []string `json:"stageParams,omitempty"`
		Mask           *bool    `json:"mask,omitempty"`
	}

	PluginScanArgs struct {
		Subscriptions  []string `json:"subscriptions,omitempty"`
		ResourceGroups []string `json:"resourceGroups,omitempty"`
		Mask           *bool    `json:"mask,omitempty"`
	}
)

func NewScanParamsWithDefaults(args ScanArgs) (*ScanParams, error) {
	stages := NewStageConfigsWithDefaults()
	if len(args.Stages) > 0 {
		if err := stages.ConfigureStages(args.Stages); err != nil {
			return nil, err
		}
	}

	if err := stages.ApplyStageParams(args.StageParams); err != nil {
		return nil, err
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
	}, nil
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
