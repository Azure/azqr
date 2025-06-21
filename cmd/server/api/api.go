// Package api provides REST API handlers for azqr supported commands.
//
// This file exposes endpoints to list supported resource types, recommendations, and to trigger scans.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// SupportedTypesHandler handles GET /api/types
// @Summary List all supported resource types
// @Description Returns details of the Azure Resource Types supported by Azure Quick Review (azqr).
// @Tags types
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/types [get]
func SupportedTypesHandler(w http.ResponseWriter, r *http.Request) {
	st := renderers.SupportedTypes{}
	output := st.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"types": output})
}

// RecommendationsHandler handles GET /api/recommendations
// @Summary List all supported recommendations
// @Description Returns details of the Recommendations supported by Azure Quick Review (azqr).
// @Tags recommendations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/recommendations [get]
func RecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	output := renderers.GetAllRecommendations(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"recommendations": output})
}

// ScanRequest represents the request body for POST /api/scan
// @Summary Run an azqr scan for a given resource type
// @Description Triggers a scan for the specified resource type
// @Tags scan
// @Accept json
// @Produce json
// @Param scan body ScanRequest true "Scan request"
// @Success 202 {object} map[string]string
// @Router /api/scan [post]
type ScanRequest struct {
	ServiceKey string `json:"key"`
}

// ScanHandler handles POST /api/scan
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	go func(serviceKey string) {
		scannerKeys := []string{serviceKey}
		filters := models.LoadFilters("", scannerKeys)
		params := internal.NewScanParams()
		params.Cost = false
		params.Defender = false
		params.Advisor = false
		params.ScannerKeys = scannerKeys
		params.Filters = filters
		scanner := internal.Scanner{}
		scanner.Scan(params)
	}(req.ServiceKey)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Scan started, please wait for the results."})
}
