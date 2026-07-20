// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package crg

import (
	"testing"
)

func TestReservationStatus(t *testing.T) {
	tests := []struct {
		name      string
		reserved  int
		allocated int
		wantStatus ReservationStatus
		wantAvail  int
	}{
		{"idle: no VMs", 10, 0, StatusIdle, 10},
		{"available: partial use", 10, 5, StatusAvailable, 5},
		{"at-capacity: fully used", 10, 10, StatusAtCapacity, 0},
		{"over-allocated: more than reserved", 10, 12, StatusOverAllocated, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := tt.reserved - tt.allocated
			if available != tt.wantAvail {
				t.Errorf("available = %d, want %d", available, tt.wantAvail)
			}

			var status ReservationStatus
			switch {
			case tt.allocated == 0:
				status = StatusIdle
			case tt.allocated > tt.reserved:
				status = StatusOverAllocated
			case available == 0:
				status = StatusAtCapacity
			default:
				status = StatusAvailable
			}

			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}
