package scanners

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// DiagnosticsSettings - analyzer
type DiagnosticsSettings struct {
	diagnosticsSettingsClient *armmonitor.DiagnosticSettingsClient
	ctx                       context.Context
	hasDiagnosticsFunc        func(resourceId string) (bool, error)
}

// Init - Initializes the DiagnosticsSettings
func (s *DiagnosticsSettings) Init(ctx context.Context, cred azcore.TokenCredential) error {
	s.ctx = ctx
	var err error
	s.diagnosticsSettingsClient, err = armmonitor.NewDiagnosticSettingsClient(cred, nil)
	if err != nil {
		return err
	}
	return nil
}

// HasDiagnostics - Checks if a resource has diagnostics settings
func (s *DiagnosticsSettings) HasDiagnostics(resourceID string) (bool, error) {
	if s.hasDiagnosticsFunc == nil {
		pager := s.diagnosticsSettingsClient.NewListPager(resourceID, nil)

		for pager.More() {
			resp, err := pager.NextPage(s.ctx)
			if err != nil {
				return false, err
			}
			if len(resp.Value) > 0 {
				return true, nil
			}
		}

		return false, nil
	}

	return s.hasDiagnosticsFunc(resourceID)
}
