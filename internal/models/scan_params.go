package models

import (
	"fmt"
)

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
	}

	Scanner struct{}
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

// Validate validates the scan parameters and returns an error if invalid
func (sp *ScanParams) Validate() error {
	if len(sp.ManagementGroups) > 0 && (len(sp.Subscriptions) > 0 || len(sp.ResourceGroups) > 0) {
		return fmt.Errorf("management Group name cannot be used with a Subscription Id or Resource Group name")
	}

	if len(sp.Subscriptions) < 1 && len(sp.ResourceGroups) > 0 {
		return fmt.Errorf("resource Group name can only be used with a Subscription Id")
	}

	if len(sp.Subscriptions) > 1 && len(sp.ResourceGroups) > 0 {
		return fmt.Errorf("resource Group name can only be used with 1 Subscription Id")
	}

	return nil
}
