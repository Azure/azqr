package analyzers

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

// NewDiagnosticsSettings - Creates a new DiagnosticsSettings
func NewDiagnosticsSettings(ctx context.Context, cred azcore.TokenCredential) (*DiagnosticsSettings, error) {
	diagnosticsSettingsClient, err := armmonitor.NewDiagnosticSettingsClient(cred, nil)
	if err != nil {
		return nil, err
	}
	settings := DiagnosticsSettings{
		diagnosticsSettingsClient: diagnosticsSettingsClient,
		ctx:                       ctx,
	}

	return &settings, nil
}

// HasDiagnostics - Checks if a resource has diagnostics settings
func (s DiagnosticsSettings) HasDiagnostics(resourceID string) (bool, error) {
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
