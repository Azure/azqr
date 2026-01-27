package models

type (
	ScanParams struct {
		ManagementGroups       []string
		Subscriptions          []string
		ResourceGroups         []string
		OutputName             string
		Defender               bool
		Advisor                bool
		Arc                    bool
		Xlsx                   bool
		Cost                   bool
		Mask                   bool
		Csv                    bool
		Json                   bool
		Stdout                 bool
		Debug                  bool
		Policy                 bool
		ScannerKeys            []string
		Filters                *Filters
		UseAzqrRecommendations bool
		UseAprlRecommendations bool
		EnabledInternalPlugins map[string]bool
		// Profiling options (only effective when built with 'debug' tag)
		CPUProfile   string
		MemProfile   string
		TraceProfile string
	}
)

func NewScanParams() *ScanParams {
	return &ScanParams{
		ManagementGroups:       []string{},
		Subscriptions:          []string{},
		ResourceGroups:         []string{},
		OutputName:             "",
		Defender:               true,
		Advisor:                true,
		Cost:                   true,
		Mask:                   true,
		Csv:                    false,
		Json:                   false,
		Debug:                  false,
		Policy:                 false,
		ScannerKeys:            []string{},
		Filters:                NewFilters(),
		UseAzqrRecommendations: true,
		UseAprlRecommendations: true,
	}
}
