// Package api provides the HTTP router for the azqr REST API.
//
// This file sets up the HTTP routes for the supported commands.
package api

import (
	"net/http"
)

// RegisterRoutes registers all REST API routes for azqr.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/types", SupportedTypesHandler)
	mux.HandleFunc("/api/recommendations", RecommendationsHandler)
	mux.HandleFunc("/api/scan", ScanHandler)
}
