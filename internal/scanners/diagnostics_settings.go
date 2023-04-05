// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// DiagnosticsSettings - analyzer
type DiagnosticsSettings struct {
	config                    *ScannerConfig
	diagnosticsSettingsClient *armmonitor.DiagnosticSettingsClient
	HasDiagnosticsFunc        func(resourceId string) (bool, error)
}

// Init - Initializes the DiagnosticsSettings
func (s *DiagnosticsSettings) Init(config *ScannerConfig) error {
	s.config = config
	var err error
	s.diagnosticsSettingsClient, err = armmonitor.NewDiagnosticSettingsClient(s.config.Cred, nil)
	if err != nil {
		return err
	}
	return nil
}

// HasDiagnostics - Checks if a resource has diagnostics settings
func (s *DiagnosticsSettings) HasDiagnostics(resourceID string) (bool, error) {
	if s.HasDiagnosticsFunc == nil {
		pager := s.diagnosticsSettingsClient.NewListPager(resourceID, nil)

		for pager.More() {
			resp, err := pager.NextPage(s.config.Ctx)
			if err != nil {
				return false, err
			}
			if len(resp.Value) > 0 {
				return true, nil
			}
		}

		return false, nil
	}

	return s.HasDiagnosticsFunc(resourceID)
}
