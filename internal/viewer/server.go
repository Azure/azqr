// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package viewer

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

//go:embed static/*
var staticFS embed.FS

// NewHandler returns an http.Handler serving API and UI.
func NewHandler(ds *DataStore) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/datasets", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, ds.ListDataSets()) })
	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, ds.Summary()) })
	mux.HandleFunc("/api/analytics", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, ds.Analytics()) })
	mux.HandleFunc("/api/plugins", func(w http.ResponseWriter, r *http.Request) {
		// Return plugin metadata (without data)
		writeJSON(w, ds.Plugins)
	})
	mux.HandleFunc("/api/plugin/", func(w http.ResponseWriter, r *http.Request) {
		pluginName := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/plugin/"), "/")
		if pluginName == "" {
			http.Error(w, "plugin name required", http.StatusBadRequest)
			return
		}
		// Find plugin and filter data
		var plugin *PluginDataset
		for i := range ds.Plugins {
			if ds.Plugins[i].Name == pluginName {
				plugin = &ds.Plugins[i]
				break
			}
		}
		if plugin == nil {
			http.Error(w, "plugin not found", http.StatusNotFound)
			return
		}
		// Filter plugin data using query parameters
		filtered := filterPluginData(plugin.Data, r.URL.Query())
		writeJSON(w, filtered)
	})
	mux.HandleFunc("/api/data/", func(w http.ResponseWriter, r *http.Request) {
		dataset := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/data/"), "/")
		if dataset == "" {
			http.Error(w, "dataset required", http.StatusBadRequest)
			return
		}
		res, err := ds.Filter(dataset, r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, res)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || p == "" {
			serveStatic(w, "static/index.html")
			return
		}
		clean := strings.TrimPrefix(path.Clean(p), "/")
		// Use path.Join instead of filepath.Join for embed.FS (always uses forward slashes)
		fp := path.Join("static", clean)
		if _, err := staticFS.Open(fp); err == nil {
			serveStatic(w, fp)
			return
		}
		serveStatic(w, "static/index.html")
	})

	return loggingMiddleware(mux)
}

// StartServer serves until context cancellation.
func StartServer(ctx context.Context, addr string, ds *DataStore) error {
	srv := &http.Server{Addr: addr, Handler: NewHandler(ds)}
	go func() {
		<-ctx.Done()
		ctxS, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxS)
	}()
	return srv.ListenAndServe()
}

// writeJSON writes JSON to the response
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// filterPluginData filters plugin data based on query parameters
func filterPluginData(data []map[string]string, query url.Values) []map[string]string {
	if len(query) == 0 {
		return data
	}

	filtered := make([]map[string]string, 0)
	for _, record := range data {
		matches := true
		for key, values := range query {
			if len(values) == 0 {
				continue
			}
			// Check if record value matches any of the filter values
			recordValue, exists := record[key]
			if !exists {
				matches = false
				break
			}
			// For search filters, check if the value contains the search term
			found := false
			for _, filterValue := range values {
				if strings.Contains(strings.ToLower(recordValue), strings.ToLower(filterValue)) {
					found = true
					break
				}
			}
			if !found {
				matches = false
				break
			}
		}
		if matches {
			filtered = append(filtered, record)
		}
	}

	return filtered
}
func serveStatic(w http.ResponseWriter, name string) {
	data, err := staticFS.ReadFile(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("file not found: %s", name), http.StatusNotFound)
		return
	}
	switch {
	case strings.HasSuffix(name, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(name, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	_, _ = w.Write(data)
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[viewer] %s %s %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}
