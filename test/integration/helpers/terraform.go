//go:build integration

package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
)

// TerraformHelper wraps Terratest's terraform module for integration tests.
type TerraformHelper struct {
	t       *testing.T
	options *terraform.Options
	dir     string
}

// NewTerraformHelper creates a new Terraform helper for the given directory
func NewTerraformHelper(t *testing.T, dir string) *TerraformHelper {
	t.Helper()

	options := &terraform.Options{
		TerraformDir: dir,
		NoColor:      false,
		Upgrade:      true, // Re-resolve providers even if lock file is missing/stale
	}

	helper := &TerraformHelper{
		t:       t,
		options: options,
		dir:     dir,
	}

	// Register cleanup to destroy resources automatically
	t.Cleanup(func() {
		helper.Destroy()
	})

	return helper
}

// Init runs terraform init, cleaning stale state first to avoid
// "inconsistent dependency lock file" errors.
func (h *TerraformHelper) Init() {
	h.t.Helper()

	// Remove stale terraform state to ensure a clean init.
	// This prevents errors when lock files or .terraform/ dirs
	// are out of sync (e.g., from partial cleanup or prior runs).
	for _, name := range []string{".terraform", ".terraform.lock.hcl"} {
		p := filepath.Join(h.dir, name)
		if err := os.RemoveAll(p); err != nil {
			h.t.Logf("Warning: failed to remove %s: %v", p, err)
		}
	}

	h.t.Logf("Running terraform init in %s", h.dir)
	terraform.Init(h.t, h.options)
	h.t.Logf("Terraform init completed successfully")
}

// Apply runs terraform apply with the given variables
func (h *TerraformHelper) Apply(vars map[string]interface{}) {
	h.t.Helper()

	if vars != nil {
		h.options.Vars = vars
	}

	h.t.Logf("Running terraform apply in %s", h.dir)
	terraform.Apply(h.t, h.options)
	h.t.Logf("Terraform apply completed successfully")
}

// InitAndApply runs both terraform init and apply
func (h *TerraformHelper) InitAndApply(vars map[string]interface{}) {
	h.t.Helper()
	h.Init()
	h.Apply(vars)
}

// Destroy runs terraform destroy
func (h *TerraformHelper) Destroy() {
	h.t.Helper()
	h.t.Logf("Running terraform destroy in %s", h.dir)
	_, err := terraform.DestroyE(h.t, h.options)
	if err != nil {
		h.t.Logf("Warning: Terraform destroy failed: %v", err)
	} else {
		h.t.Logf("Terraform destroy completed successfully")
	}
}

// GetOutput retrieves a terraform output value
func (h *TerraformHelper) GetOutput(key string) string {
	h.t.Helper()
	return terraform.Output(h.t, h.options, key)
}

// GetOutputMap retrieves all terraform outputs as a map
func (h *TerraformHelper) GetOutputMap() map[string]interface{} {
	h.t.Helper()
	return terraform.OutputAll(h.t, h.options)
}

// RequireOutput asserts that an output exists and returns its value
func (h *TerraformHelper) RequireOutput(key string) string {
	h.t.Helper()
	value := h.GetOutput(key)
	require.NotEmpty(h.t, value, "Terraform output '%s' should not be empty", key)
	return value
}

// SetVariables sets terraform variables
func (h *TerraformHelper) SetVariables(vars map[string]interface{}) {
	h.t.Helper()
	h.options.Vars = vars
}

// SetEnvVars sets environment variables for terraform execution
func (h *TerraformHelper) SetEnvVars(envVars map[string]string) {
	h.t.Helper()
	h.options.EnvVars = envVars
}
