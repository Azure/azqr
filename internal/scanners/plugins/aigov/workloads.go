// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "fmt"

// defaultWorkloadID mirrors TokenLens's default_technical_workload: absent
// any business-provided workload metadata (which azqr does not currently
// collect), every deployment is its own technical workload so cost is always
// visible.
func defaultWorkloadID(accountName, deploymentName string) string {
	return fmt.Sprintf("%s/%s", accountName, deploymentName)
}
