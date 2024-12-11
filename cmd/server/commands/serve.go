// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var globalDB *sql.DB

func init() {
	serveCmd.PersistentFlags().Int("port", 8080, "Port to listen to")

	rootCmd.AddCommand(serveCmd)
}

// initDatabase initializes the SQLite database and creates the scans table.
func initDatabase() (*sql.DB, error) {
	// Define the path for the .azqr folder and database file
	azqrFolder := ".azqr"
	dbFilePath := fmt.Sprintf("%s/azqr_scans.db", azqrFolder)

	// Create the .azqr folder if it doesn't exist
	err := os.MkdirAll(azqrFolder, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create .azqr folder: %w", err)
	}

	// Open the SQLite database file (or create it if it doesn't exist)
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the scans table if it doesn't exist
	createTableQuery := `CREATE TABLE IF NOT EXISTS scans (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		parameters TEXT,
		result TEXT
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create scans table: %w", err)
	}

	return db, nil
}

func SetDatabase(db *sql.DB) {
	globalDB = db
}

// StatefulScanner encapsulates the business logic for handling scan operations.
type StatefulScanner struct {
	DB *sql.DB
}

// NewStatefulScanner creates a new instance of DomainService.
func NewStatefulScanner(db *sql.DB) *StatefulScanner {
	return &StatefulScanner{DB: db}
}

// RunScan initiates a new scan and stores it in the database.
func (ds *StatefulScanner) RunScan(params *internal.ScanParams) (string, error) {
	// Generate a unique scan ID
	scanID := uuid.New().String()

	// Insert the scan record into the database
	insertQuery := `INSERT INTO scans (id, status, parameters) VALUES (?, ?, ?)`
	_, err := ds.DB.Exec(insertQuery, scanID, "Pending", "")
	if err != nil {
		return "", fmt.Errorf("failed to store scan record: %w", err)
	}

	// Run the scan asynchronously
	go func() {
		scannerKeys, _ := models.GetScanners()
		filters := models.LoadFilters("", scannerKeys)
		params.ScannerKeys = scannerKeys
		params.Filters = filters
		scanner := internal.Scanner{}
		scanner.Scan(params)

		// Update the scan status to Completed (example logic)
		updateQuery := `UPDATE scans SET status = ?, result = ? WHERE id = ?`
		_, _ = ds.DB.Exec(updateQuery, "Completed", "{}", scanID) // Replace "{}" with actual results
	}()

	return scanID, nil
}

// GetScanResult retrieves the result of a scan by its ID.
func (ds *StatefulScanner) GetScanResult(scanID string) (map[string]string, error) {
	query := `SELECT status, result FROM scans WHERE id = ?`
	row := ds.DB.QueryRow(query, scanID)

	var status, result string
	err := row.Scan(&status, &result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan ID not found")
		}
		return nil, fmt.Errorf("failed to retrieve scan record: %w", err)
	}

	return map[string]string{
		"scanID": scanID,
		"status": status,
		"result": result,
	}, nil
}

// GetAllScans retrieves all scan records from the database.
func (ds *StatefulScanner) GetAllScans() ([]map[string]interface{}, error) {
	query := `SELECT id, status, createdAt, updatedAt, parameters, result FROM scans`
	rows, err := ds.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve scans: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Fatal().Err(err).Msg("Failed to close rows")
		}
	}()

	var scans []map[string]interface{}
	for rows.Next() {
		var id, status, createdAt, updatedAt, parameters, result string
		if err := rows.Scan(&id, &status, &createdAt, &updatedAt, &parameters, &result); err != nil {
			return nil, fmt.Errorf("failed to parse scan record: %w", err)
		}
		scans = append(scans, map[string]interface{}{
			"id":         id,
			"status":     status,
			"createdAt":  createdAt,
			"updatedAt":  updatedAt,
			"parameters": parameters,
			"result":     result,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scan records: %w", err)
	}

	return scans, nil
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  "Start the API server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd)
	},
}

// Add a new endpoint to trigger a scan
func runScanHandler(w http.ResponseWriter, r *http.Request) {
	params := internal.NewScanParams()
	domainService := NewStatefulScanner(globalDB)

	scanID, err := domainService.RunScan(params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to initiate scan")
		return
	}

	response := map[string]string{
		"scanID":  scanID,
		"message": "Scan initiated successfully",
	}
	respondWithJSON(w, http.StatusAccepted, response)
}

// Add a new endpoint to retrieve scan results by ID
func getScanResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["id"]

	domainService := NewStatefulScanner(globalDB)
	result, err := domainService.GetScanResult(scanID)
	if err != nil {
		if err.Error() == "scan ID not found" {
			respondWithError(w, http.StatusNotFound, "Scan ID not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve scan record")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// getAllScansHandler handles the request to retrieve all scans.
// It fetches all scan records from the database and returns them as JSON.
func getAllScansHandler(w http.ResponseWriter, r *http.Request) {
	domainService := NewStatefulScanner(globalDB)
	scans, err := domainService.GetAllScans()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve scans")
		return
	}

	respondWithJSON(w, http.StatusOK, scans)
}

// Refactor the serve function to initialize the database when the server starts
func serve(cmd *cobra.Command) {
	// Initialize the database once when the server starts
	db, err := initDatabase()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal().Err(err).Msg("Failed to close database")
		}
	}()

	// Set the global database instance
	SetDatabase(db)

	// Get the port from command flags
	port, _ := cmd.Flags().GetInt("port")

	// Create a new router
	r := mux.NewRouter()
	api := r.PathPrefix("/api/").Subrouter()

	// Register scan-related endpoints
	api.HandleFunc("/scan", runScanHandler).Methods("POST")
	api.HandleFunc("/scan/{id}", getScanResultHandler).Methods("GET")
	api.HandleFunc("/scans", getAllScansHandler).Methods("GET")

	// Start the HTTP server
	http.Handle("/", r)
	log.Info().Msgf("Starting server on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
