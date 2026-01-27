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
)

func NewScanParams() *ScanParams {
	return &ScanParams{
		ManagementGroups: []string{},
		Subscriptions:    []string{},
		ResourceGroups:   []string{},
		OutputName:       "",
		Stages:           NewStageConfigsWithDefaults(),
		Mask:             true,
		Csv:              false,
		Json:             false,
		Debug:            false,
		ScannerKeys:      []string{},
		Filters:          NewFilters(),
	}
}
