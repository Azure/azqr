package analyzers

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

type DiagnosticsSettings struct {
	diagnosticsSettingsClient *armmonitor.DiagnosticSettingsClient
	ctx                       context.Context
	hasDiagnosticsFunc        func(resourceId string) (bool, error)
}

func NewDiagnosticsSettings(cred azcore.TokenCredential, ctx context.Context) (*DiagnosticsSettings, error) {
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

func (s DiagnosticsSettings) HasDiagnostics(resourceId string) (bool, error) {
	if s.hasDiagnosticsFunc == nil {
		pager := s.diagnosticsSettingsClient.NewListPager(resourceId, nil)

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
	} else {
		return s.hasDiagnosticsFunc(resourceId)
	}
}
