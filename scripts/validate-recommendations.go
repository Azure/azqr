package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ValidationError struct {
	File    string
	Rule    int
	Field   string
	Message string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-recommendations-dir> [<path-to-recommendations-dir>...]\n", os.Args[0])
		os.Exit(1)
	}

	allErrors := []ValidationError{}

	// Load the JSON schema once (relative to the script location)
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	repoRoot := filepath.Join(filepath.Dir(execPath), "..")
	schemaPath := filepath.Join(repoRoot, "internal", "graph", "schema", "recommendations.schema.json")

	// If running with 'go run', adjust the path
	if !fileExists(schemaPath) {
		// When using 'go run', we're in a temp directory, so use current working directory
		cwd, _ := os.Getwd()
		schemaPath = filepath.Join(cwd, "internal", "graph", "schema", "recommendations.schema.json")
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)

	// Process each directory argument
	for i := 1; i < len(os.Args); i++ {
		dir := os.Args[i]

		// Convert to absolute path
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting absolute path for %s: %v\n", dir, err)
			os.Exit(1)
		}

		err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
				fileErrors := validateFile(path, schemaLoader)
				allErrors = append(allErrors, fileErrors...)
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	if len(allErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Validation failed with %d error(s):\n\n", len(allErrors))
		for _, e := range allErrors {
			if e.Rule > 0 {
				fmt.Fprintf(os.Stderr, "  %s [Rule #%d] %s: %s\n", e.File, e.Rule, e.Field, e.Message)
			} else {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", e.File, e.Message)
			}
		}
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	fmt.Println("✅ All recommendation files are valid!")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateFile(path string, schemaLoader gojsonschema.JSONLoader) []ValidationError {
	errors := []ValidationError{}

	// Read YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		errors = append(errors, ValidationError{
			File:    path,
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return errors
	}

	// Parse YAML
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		errors = append(errors, ValidationError{
			File:    path,
			Message: fmt.Sprintf("Failed to parse YAML: %v", err),
		})
		return errors
	}

	// Convert to JSON for schema validation
	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		errors = append(errors, ValidationError{
			File:    path,
			Message: fmt.Sprintf("Failed to convert to JSON: %v", err),
		})
		return errors
	}

	// Validate against schema
	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		errors = append(errors, ValidationError{
			File:    path,
			Message: fmt.Sprintf("Schema validation error: %v", err),
		})
		return errors
	}

	// Process validation results
	if !result.Valid() {
		// Try to parse as array to get rule numbers
		var recommendations []map[string]interface{}
		_ = json.Unmarshal(jsonData, &recommendations)

		for _, resultErr := range result.Errors() {
			field := resultErr.Field()
			context := resultErr.Context().String()
			description := resultErr.Description()

			// Extract rule number from context path like "(root).0.aprlGuid"
			ruleNum := 0
			if strings.Contains(context, "(root).") {
				parts := strings.Split(context, ".")
				if len(parts) > 1 {
					_, _ = fmt.Sscanf(parts[1], "%d", &ruleNum)
					ruleNum++ // Convert 0-indexed to 1-indexed
				}
			}

			// Clean up field name
			fieldName := strings.TrimPrefix(field, "(root).")
			if idx := strings.Index(fieldName, "."); idx > 0 {
				fieldName = fieldName[idx+1:]
			}

			errors = append(errors, ValidationError{
				File:    path,
				Rule:    ruleNum,
				Field:   fieldName,
				Message: description,
			})
		}
	}

	return errors
}
