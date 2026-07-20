// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package crg fetches Azure Capacity Reservation Group inventory for migration planning.
// It uses the armcompute SDK clients:
//
//  1. List CRGs:         CapacityReservationGroupsClient.NewListBySubscriptionPager
//  2. List reservations: CapacityReservationsClient.NewListByCapacityReservationGroupPager (with instanceView expand)
package crg

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/rs/zerolog/log"
)

// ReservationStatus classifies the utilisation of one capacity reservation.
type ReservationStatus string

const (
	StatusIdle          ReservationStatus = "Idle"           // reserved but no VMs allocated
	StatusAvailable     ReservationStatus = "Available"      // some capacity in use, headroom remains
	StatusAtCapacity    ReservationStatus = "At-Capacity"    // fully utilised, no free slots
	StatusOverAllocated ReservationStatus = "Over-Allocated" // allocated instances exceed reserved count
)

// ReservationEntry is one capacity reservation inside a CRG.
type ReservationEntry struct {
	SubscriptionID   string
	SubscriptionName string
	ResourceGroup    string
	CRGName          string
	ReservationName  string
	Location         string
	SKU              string
	Reserved         int // sku.capacity
	Allocated        int // len(instanceView.utilizationInfo.virtualMachinesAllocated)
	Available        int // Reserved - Allocated
	Status           ReservationStatus
}

// ─── Fetch ───────────────────────────────────────────────────────────────────

// FetchReservations returns all capacity reservations for a subscription.
// Returns an empty slice (not an error) when the subscription has no CRGs.
func FetchReservations(ctx context.Context, cred azcore.TokenCredential, clientOpts *arm.ClientOptions, subscriptionID, subscriptionName string) ([]ReservationEntry, error) {
	crgClient, err := armcompute.NewCapacityReservationGroupsClient(subscriptionID, cred, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRG client for %s: %w", subscriptionID, err)
	}
	crClient, err := armcompute.NewCapacityReservationsClient(subscriptionID, cred, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create capacity reservations client for %s: %w", subscriptionID, err)
	}

	var entries []ReservationEntry
	pager := crgClient.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("CRG list error for %s: %w", subscriptionID, err)
		}
		for _, crgItem := range page.Value {
			if crgItem.ID == nil || crgItem.Name == nil {
				continue
			}
			rg := models.GetResourceGroupFromResourceID(*crgItem.ID)
			log.Debug().Msgf("Fetching reservations for CRG %s/%s", rg, *crgItem.Name)

			crPager := crClient.NewListByCapacityReservationGroupPager(rg, *crgItem.Name, nil)
			for crPager.More() {
				crPage, err := crPager.NextPage(ctx)
				if err != nil {
					log.Warn().Err(err).Msgf("Skipping reservations for CRG %s: API error", *crgItem.Name)
					break
				}
				for _, crSummary := range crPage.Value {
					if crSummary.Name == nil {
						continue
					}
					// Fetch instance view separately to get VirtualMachinesAllocated utilization data.
					cr, err := crClient.Get(ctx, rg, *crgItem.Name, *crSummary.Name,
						&armcompute.CapacityReservationsClientGetOptions{
							Expand: to.Ptr(armcompute.CapacityReservationInstanceViewTypesInstanceView),
						})
					if err != nil {
						log.Warn().Err(err).Msgf("Skipping reservation %s in CRG %s: Get error", *crSummary.Name, *crgItem.Name)
						continue
					}
					reserved := 0
				if cr.SKU != nil && cr.SKU.Capacity != nil {
						reserved = int(*cr.SKU.Capacity)
					}

					allocated := 0
					if cr.Properties != nil &&
						cr.Properties.InstanceView != nil &&
						cr.Properties.InstanceView.UtilizationInfo != nil {
						allocated = len(cr.Properties.InstanceView.UtilizationInfo.VirtualMachinesAllocated)
					}
					available := reserved - allocated

					var status ReservationStatus
					switch {
					case allocated == 0:
						status = StatusIdle
					case available < 0:
						status = StatusOverAllocated
					case available == 0:
						status = StatusAtCapacity
					default:
						status = StatusAvailable
					}

					location := ""
					if crgItem.Location != nil {
						location = strings.ToLower(*crgItem.Location)
					}
					crName := ""
					if cr.Name != nil {
						crName = *cr.Name
					}
					skuName := ""
					if cr.SKU != nil && cr.SKU.Name != nil {
						skuName = *cr.SKU.Name
					}

					entries = append(entries, ReservationEntry{
						SubscriptionID:   subscriptionID,
						SubscriptionName: subscriptionName,
						ResourceGroup:    rg,
						CRGName:          *crgItem.Name,
						ReservationName:  crName,
						Location:         location,
						SKU:              skuName,
						Reserved:         reserved,
						Allocated:        allocated,
						Available:        available,
						Status:           status,
					})
				}
			}
		}
	}

	log.Debug().Msgf("Fetched %d capacity reservations for subscription %s", len(entries), subscriptionID)
	return entries, nil
}
