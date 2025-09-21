package viewer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var sample = &DataStore{Data: map[string][]map[string]string{
	DataSetRecommendations:         {{"implemented": "true", "recommendationId": "r1"}, {"implemented": "false", "recommendationId": "r2"}},
	DataSetImpacted:                {{"recommendationId": "r2", "resourceId": "res1"}},
	DataSetResourceType:            {{"resourceType": "Microsoft.Compute/virtualMachines"}},
	DataSetInventory:               {{"resourceType": "Microsoft.Compute/virtualMachines", "resourceName": "vm1"}},
	DataSetAdvisor:                 {},
	DataSetAzurePolicy:             {{"complianceState": "NonCompliant"}},
	DataSetDefender:                {},
	DataSetDefenderRecommendations: {},
	DataSetCosts:                   {{"value": "10.50"}},
	DataSetOutOfScope:              {},
}}

func TestSummaryEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewHandler(sample))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/summary")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestFilterEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewHandler(sample))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/data/recommendations?implemented=true")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestAnalyticsEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewHandler(sample))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/analytics")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}
