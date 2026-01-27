// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"net"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "azqr" {
		t.Errorf("Expected rootCmd.Use to be 'azqr', got %q", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	expectedCommands := []string{"scan", "compare", "show", "rules", "types", "plugins"}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q not found in root command", expectedCmd)
		}
	}
}

func TestScanCommandExists(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
		return
	}

	if scanCmd.Use != "scan" {
		t.Errorf("Expected scan command Use to be 'scan', got %q", scanCmd.Use)
	}

	if scanCmd.Short == "" {
		t.Error("scan command Short should not be empty")
	}
}

func TestScanCommandHasRequiredFlags(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
		return
	}

	requiredFlags := []struct {
		name         string
		expectedType string
	}{
		{"management-group-id", "stringArray"},
		{"subscription-id", "stringArray"},
		{"resource-group", "stringArray"},
		{"stages", "stringArray"},
		{"plugin", "stringArray"},
		{"xlsx", "bool"},
		{"json", "bool"},
		{"csv", "bool"},
		{"output-name", "string"},
		{"mask", "bool"},
		{"filters", "string"},
	}

	for _, rf := range requiredFlags {
		// Check persistent flags (scan command uses PersistentFlags)
		flag := scanCmd.PersistentFlags().Lookup(rf.name)
		if flag == nil {
			t.Errorf("Expected flag %q not found in scan command", rf.name)
			continue
		}

		if flag.Value.Type() != rf.expectedType {
			t.Errorf("Flag %q: expected type %q, got %q", rf.name, rf.expectedType, flag.Value.Type())
		}
	}
}

func TestScanCommandFlagDefaults(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
	}

	tests := []struct {
		flagName      string
		expectedValue string
	}{
		{"xlsx", "true"},
		{"json", "false"},
		{"csv", "false"},
		{"mask", "true"},
	}

	for _, tt := range tests {
		// Check persistent flags
		flag := scanCmd.PersistentFlags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("Flag %q not found", tt.flagName)
			continue
		}

		if flag.DefValue != tt.expectedValue {
			t.Errorf("Flag %q: expected default value %q, got %q", tt.flagName, tt.expectedValue, flag.DefValue)
		}
	}
}

func TestServiceCommandsAreRegistered(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
	}

	// Test a sample of service commands
	expectedServiceCommands := []string{
		"aa", "adf", "afd", "afw", "agw", "aks", "apim",
		"arc", "asp", "cosmos", "kv", "sql", "st", "vnet",
	}

	for _, serviceName := range expectedServiceCommands {
		found := false
		for _, cmd := range scanCmd.Commands() {
			if cmd.Name() == serviceName {
				found = true
				// Verify the command has proper metadata
				if cmd.Short == "" {
					t.Errorf("Service command %q has empty Short description", serviceName)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected service command %q not found under scan", serviceName)
		}
	}
}

func TestCompareCommandExists(t *testing.T) {
	compareCmd := findCommand(rootCmd, "compare")
	if compareCmd == nil {
		t.Fatal("compare command should exist")
		return
	}

	if compareCmd.Use != "compare" {
		t.Errorf("Expected compare command Use to be 'compare', got %q", compareCmd.Use)
	}

	if compareCmd.Short == "" {
		t.Error("compare command Short should not be empty")
	}
}

func TestCompareCommandHasRequiredFlags(t *testing.T) {
	compareCmd := findCommand(rootCmd, "compare")
	if compareCmd == nil {
		t.Fatal("compare command should exist")
		return
	}

	requiredFlags := []string{"file1", "file2"}
	for _, flagName := range requiredFlags {
		flag := compareCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found in compare command", flagName)
			continue
		}
	}

	// Check optional flags
	optionalFlags := []string{"format", "output"}
	for _, flagName := range optionalFlags {
		flag := compareCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %q not found in compare command", flagName)
		}
	}
}

func TestShowCommandExists(t *testing.T) {
	showCmd := findCommand(rootCmd, "show")
	if showCmd == nil {
		t.Fatal("show command should exist")
		return
	}

	if showCmd.Use != "show" {
		t.Errorf("Expected show command Use to be 'show', got %q", showCmd.Use)
	}

	// Check required flags
	fileFlag := showCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("show command should have 'file' flag")
	}

	portFlag := showCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("show command should have 'port' flag")
	}

	openFlag := showCmd.Flags().Lookup("open")
	if openFlag == nil {
		t.Error("show command should have 'open' flag")
	}
}

func TestRulesCommandExists(t *testing.T) {
	rulesCmd := findCommand(rootCmd, "rules")
	if rulesCmd == nil {
		t.Fatal("rules command should exist")
		return
	}

	if rulesCmd.Use != "rules" {
		t.Errorf("Expected rules command Use to be 'rules', got %q", rulesCmd.Use)
	}

	// Check for json flag
	jsonFlag := rootCmd.PersistentFlags().Lookup("json")
	if jsonFlag == nil {
		t.Error("rules command should have 'json' flag")
	}
}

func TestTypesCommandExists(t *testing.T) {
	typesCmd := findCommand(rootCmd, "types")
	if typesCmd == nil {
		t.Fatal("types command should exist")
		return
	}

	if typesCmd.Use != "types" {
		t.Errorf("Expected types command Use to be 'types', got %q", typesCmd.Use)
	}

	if typesCmd.Short == "" {
		t.Error("types command Short should not be empty")
	}
}

func TestPluginsCommandExists(t *testing.T) {
	pluginsCmd := findCommand(rootCmd, "plugins")
	if pluginsCmd == nil {
		t.Fatal("plugins command should exist")
	}

	// Check for list and info subcommands
	listCmd := findCommand(pluginsCmd, "list")
	if listCmd == nil {
		t.Error("plugins command should have 'list' subcommand")
	}

	infoCmd := findCommand(pluginsCmd, "info")
	if infoCmd == nil {
		t.Error("plugins command should have 'info' subcommand")
	}
}

func TestPortAvailable(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "high port number",
			port:    59999,
			wantErr: false,
		},
		{
			name:    "another high port",
			port:    58888,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := portAvailable(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("portAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPortAvailableWithUsedPort(t *testing.T) {
	// Start a listener on a port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to start test listener: %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	// Get the port that's in use
	addr := listener.Addr().(*net.TCPAddr)
	usedPort := addr.Port

	// Test that portAvailable detects the used port
	err = portAvailable(usedPort)
	if err == nil {
		t.Error("portAvailable() should return error for port in use")
	}
}

func TestCommandArgsValidation(t *testing.T) {
	tests := []struct {
		commandName  string
		expectNoArgs bool
	}{
		{"scan", true},
		{"compare", true},
		{"show", true},
		{"rules", true},
		{"types", true},
	}

	for _, tt := range tests {
		t.Run(tt.commandName, func(t *testing.T) {
			cmd := findCommand(rootCmd, tt.commandName)
			if cmd == nil {
				t.Fatalf("Command %q not found", tt.commandName)
			}

			if tt.expectNoArgs && cmd.Args == nil {
				t.Errorf("Command %q should have Args validator set to NoArgs", tt.commandName)
			}
		})
	}
}

func TestVersionIsSet(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version should be set")
	}
}

// Helper function to find a command by name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestOpenBrowserFunction(t *testing.T) {
	// Test that openBrowser doesn't panic with valid URL
	// We can't actually test browser opening in CI, but we can verify the function exists
	// Note: openBrowser may fail in CI environment, so we don't fail the test
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "http URL",
			url:  "http://localhost:8080",
		},
		{
			name: "https URL",
			url:  "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the function doesn't panic
			// The actual browser opening is OS-dependent and may fail in CI
			err := openBrowser(tt.url)
			// We don't assert on error because it's expected to fail in headless environments
			t.Logf("openBrowser(%s) returned: %v", tt.url, err)
		})
	}
}

func TestCompareExcelFilesFunction(t *testing.T) {
	// This function requires actual Excel files to test properly
	// We'll test error cases only

	tests := []struct {
		name    string
		file1   string
		file2   string
		wantErr bool
	}{
		{
			name:    "non-existent files",
			file1:   "/tmp/nonexistent1.xlsx",
			file2:   "/tmp/nonexistent2.xlsx",
			wantErr: true,
		},
		{
			name:    "empty file paths",
			file1:   "",
			file2:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compareExcelFiles(tt.file1, tt.file2)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareExcelFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAllServiceCommandsHaveMetadata(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
	}

	// Check all service subcommands have proper metadata
	for _, cmd := range scanCmd.Commands() {
		if cmd.Use == "" {
			t.Errorf("Service command has empty Use field")
		}
		if cmd.Short == "" {
			t.Errorf("Service command %q has empty Short description", cmd.Use)
		}
		// Long description is optional but should be valid if present
		if cmd.Long == "" {
			t.Logf("Note: Service command %q has empty Long description", cmd.Use)
		}
	}
}

func TestScanCommandShortcutFlags(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
	}

	// Test that shortcut flags are properly defined
	tests := []struct {
		flagName string
		shortcut string
	}{
		{"subscription-id", "s"},
		{"resource-group", "g"},
		{"output-name", "o"},
		{"mask", "m"},
		{"filters", "e"},
	}

	for _, tt := range tests {
		// Check persistent flags
		flag := scanCmd.PersistentFlags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("Flag %q not found", tt.flagName)
			continue
		}

		if flag.Shorthand != tt.shortcut {
			t.Errorf("Flag %q: expected shorthand %q, got %q",
				tt.flagName, tt.shortcut, flag.Shorthand)
		}
	}
}

func TestCompareCommandFormatDefault(t *testing.T) {
	compareCmd := findCommand(rootCmd, "compare")
	if compareCmd == nil {
		t.Fatal("compare command should exist")
	}

	formatFlag := compareCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("compare command should have 'format' flag")
		return
	}

	if formatFlag.DefValue != "excel" {
		t.Errorf("format flag: expected default value 'excel', got %q", formatFlag.DefValue)
	}
}

func TestShowCommandPortDefault(t *testing.T) {
	showCmd := findCommand(rootCmd, "show")
	if showCmd == nil {
		t.Fatal("show command should exist")
	}

	portFlag := showCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("show command should have 'port' flag")
		return
	}

	if portFlag.DefValue != "8080" {
		t.Errorf("port flag: expected default value '8080', got %q", portFlag.DefValue)
	}

	openFlag := showCmd.Flags().Lookup("open")
	if openFlag == nil {
		t.Fatal("show command should have 'open' flag")
		return
	}

	if openFlag.DefValue != "true" {
		t.Errorf("open flag: expected default value 'true', got %q", openFlag.DefValue)
	}
}
