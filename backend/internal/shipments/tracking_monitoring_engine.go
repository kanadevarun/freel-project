package shipments

import (
	"context"
	"fmt"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// SyncTrackingAlerts reconciles calculated intelligence alerts with persisted alert lifecycle records.
// Respects existing ACKNOWLEDGED and SUPPRESSED states, automatically resolves cleared conditions,
// and logs operator audit activity records.
func SyncTrackingAlerts(
	ctx context.Context,
	repo Repository,
	sh *spec.Shipment,
	calculatedAlerts []spec.TrackingAlert,
	actor string,
) ([]*spec.ShipmentTrackingAlertRecord, error) {
	now := time.Now()
	existingAlerts, err := repo.GetTrackingAlerts(ctx, sh.OrgID, sh.ID, "ALL")
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]*spec.ShipmentTrackingAlertRecord)
	for _, ea := range existingAlerts {
		existingMap[ea.AlertKey] = ea
	}

	calcMap := make(map[string]spec.TrackingAlert)
	for _, ca := range calculatedAlerts {
		calcMap[ca.ID] = ca
	}

	// 1. Process active calculated alerts
	for _, ca := range calculatedAlerts {
		ea, exists := existingMap[ca.ID]
		desc := ca.Description

		if !exists {
			// New alert detected -> Persist as OPEN
			newAlert := &spec.ShipmentTrackingAlertRecord{
				OrgID:             sh.OrgID,
				ShipmentID:        sh.ID,
				AlertKey:          ca.ID,
				AlertType:         ca.Type,
				Severity:          ca.Severity,
				Title:             ca.Title,
				Description:       &desc,
				Status:            spec.TrackingAlertStatusOpen,
				FirstDetectedAt:   ca.CreatedAt,
				LastDetectedAt:    now,
				NotificationCount: 1,
				LastNotifiedAt:    &now,
			}
			if err := repo.CreateTrackingAlert(ctx, newAlert); err == nil {
				_ = repo.CreateActivity(
					ctx,
					sh.OrgID,
					"SHIPMENT",
					sh.ID,
					spec.SHIPMENT_TRACKING_ALERT_DETECTED,
					fmt.Sprintf("Tracking alert detected: %s (%s)", ca.Title, ca.Severity),
					actor,
				)
			}
		} else {
			// Existing alert still present
			ea.LastDetectedAt = now
			ea.Severity = ca.Severity
			ea.Title = ca.Title
			ea.Description = &desc

			// If it was previously RESOLVED but condition returned, reopen to OPEN
			if ea.Status == spec.TrackingAlertStatusResolved {
				ea.Status = spec.TrackingAlertStatusOpen
				ea.ResolvedAt = nil
				ea.ResolvedBy = nil
				ea.FirstDetectedAt = now
				_ = repo.CreateActivity(
					ctx,
					sh.OrgID,
					"SHIPMENT",
					sh.ID,
					spec.SHIPMENT_TRACKING_ALERT_DETECTED,
					fmt.Sprintf("Tracking alert re-opened: %s", ca.Title),
					actor,
				)
			}
			// Important: If SUPPRESSED or ACKNOWLEDGED, keep the status intact!

			_ = repo.UpdateTrackingAlert(ctx, ea)
		}
	}

	// 2. Auto-resolve alerts whose calculated conditions are no longer present
	for _, ea := range existingAlerts {
		if _, stillActive := calcMap[ea.AlertKey]; !stillActive {
			if ea.Status == spec.TrackingAlertStatusOpen || ea.Status == spec.TrackingAlertStatusAcknowledged {
				ea.Status = spec.TrackingAlertStatusResolved
				ea.ResolvedAt = &now
				_ = repo.UpdateTrackingAlert(ctx, ea)
				_ = repo.CreateActivity(
					ctx,
					sh.OrgID,
					"SHIPMENT",
					sh.ID,
					spec.SHIPMENT_TRACKING_ALERT_RESOLVED,
					fmt.Sprintf("Tracking alert auto-resolved: %s", ea.Title),
					"System Engine",
				)
			}
		}
	}

	// Return updated persisted alerts
	return repo.GetTrackingAlerts(ctx, sh.OrgID, sh.ID, "ALL")
}

// BuildTrackingMonitoringSummary aggregates monitoring metrics, freshness, and recommended actions.
func BuildTrackingMonitoringSummary(
	ctx context.Context,
	repo Repository,
	orgID int64,
	shipmentID int64,
	freshness string,
	lastRefresh *time.Time,
) (*spec.TrackingMonitoringSummary, error) {
	summary, err := repo.GetTrackingMonitoringSummary(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	summary.TrackingFreshness = freshness
	summary.TrackingProvider = "Satellite AIS Telemetry Engine"
	summary.LastTrackingRefresh = lastRefresh

	// Derive next recommended operational action
	if summary.CriticalAlerts > 0 {
		summary.NextRecommendedAction = "Review and resolve active critical operational risk in Exceptions panel."
	} else if summary.OpenAlerts > 0 {
		summary.NextRecommendedAction = "Acknowledge pending schedule variance and inform consignee."
	} else if freshness == spec.TrackingFreshnessStale {
		summary.NextRecommendedAction = "Trigger manual satellite AIS refresh to update vessel coordinates."
	} else {
		summary.NextRecommendedAction = "All monitoring checks optimal. Continue standard operational tracking."
	}

	return summary, nil
}
