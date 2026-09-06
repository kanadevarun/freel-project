package shipments

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// TrackingRefreshScheduler manages automated, state-aware background telemetry refreshes (Task 17.7)
type TrackingRefreshScheduler struct {
	svc        Service
	repo       Repository
	config     spec.TrackingRefreshConfig
	activeJobs sync.Map // shipmentID -> bool for in-flight concurrency protection
	stopCh     chan struct{}
	running    bool
	mu         sync.Mutex
}

// NewTrackingRefreshScheduler initializes the tracking scheduler with environment configuration
func NewTrackingRefreshScheduler(svc Service, repo Repository) *TrackingRefreshScheduler {
	autoRefreshStr := strings.ToLower(strings.TrimSpace(os.Getenv("TRACKING_AUTO_REFRESH_ENABLED")))
	autoRefreshEnabled := autoRefreshStr != "false"

	bookedMin := parseEnvInt("TRACKING_REFRESH_INTERVAL_BOOKED", 360)
	departedMin := parseEnvInt("TRACKING_REFRESH_INTERVAL_DEPARTED", 120)
	inTransitMin := parseEnvInt("TRACKING_REFRESH_INTERVAL_IN_TRANSIT", 60)
	arrivedMin := parseEnvInt("TRACKING_REFRESH_INTERVAL_ARRIVED", 180)
	checkIntervalSec := parseEnvInt("TRACKING_REFRESH_CHECK_INTERVAL_SECONDS", 60)

	cfg := spec.TrackingRefreshConfig{
		AutoRefreshEnabled: autoRefreshEnabled,
		IntervalBooked:     time.Duration(bookedMin) * time.Minute,
		IntervalDeparted:   time.Duration(departedMin) * time.Minute,
		IntervalInTransit:  time.Duration(inTransitMin) * time.Minute,
		IntervalArrived:    time.Duration(arrivedMin) * time.Minute,
		CheckInterval:      time.Duration(checkIntervalSec) * time.Second,
	}

	return &TrackingRefreshScheduler{
		svc:    svc,
		repo:   repo,
		config: cfg,
		stopCh: make(chan struct{}),
	}
}

// Start launches the background scheduler loop if enabled
func (s *TrackingRefreshScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	if !s.config.AutoRefreshEnabled {
		log.Println("🛰️ Tracking Background Refresh Scheduler: DISABLED (TRACKING_AUTO_REFRESH_ENABLED=false)")
		return
	}

	log.Printf("🛰️ Tracking Background Refresh Scheduler: STARTED (Check Interval: %v)", s.config.CheckInterval)

	go func() {
		ticker := time.NewTicker(s.config.CheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.Stop()
				return
			case <-s.stopCh:
				log.Println("🛰️ Tracking Background Refresh Scheduler: STOPPED")
				return
			case <-ticker.C:
				cycleCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				if err := s.RunCycle(cycleCtx); err != nil {
					log.Printf("🛰️ Tracking Refresh Cycle Warning: %v", err)
				}
				cancel()
			}
		}
	}()
}

// Stop gracefully stops the background scheduler
func (s *TrackingRefreshScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

// RunCycle evaluates all active shipments and executes due refreshes
func (s *TrackingRefreshScheduler) RunCycle(ctx context.Context) error {
	if !s.config.AutoRefreshEnabled {
		return nil
	}

	activeShipments, err := s.repo.GetActiveShipmentsForRefresh(ctx)
	if err != nil {
		return err
	}

	for _, sh := range activeShipments {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 1. Evaluate whether shipment is eligible & due based on status
		interval := s.getIntervalForStatus(sh.Status)
		if interval <= 0 {
			continue // Delivered, closed, or inactive
		}

		// 2. Check latest refresh timestamp
		latestRun, err := s.repo.GetLatestTrackingRefresh(ctx, sh.OrgID, sh.ID)
		if err == nil && latestRun != nil {
			if time.Since(latestRun.StartedAt) < interval {
				continue // Not due yet
			}
		}

		// 3. Concurrency Protection: Ensure only one job per shipment at a time
		if _, loaded := s.activeJobs.LoadOrStore(sh.ID, true); loaded {
			continue // Already in-flight
		}

		// 4. Execute refresh in background worker without blocking entire cycle
		go func(targetSh *spec.Shipment) {
			defer s.activeJobs.Delete(targetSh.ID)

			jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := s.svc.RefreshShipmentTracking(
				jobCtx,
				targetSh.OrgID,
				targetSh.ID,
				nil,
				"Background Refresh Engine",
			)
			if err != nil {
				log.Printf("🛰️ [Tracking Scheduler] Shipment #%d automated refresh error: %v", targetSh.ID, err)
			}
		}(sh)
	}

	return nil
}

func (s *TrackingRefreshScheduler) getIntervalForStatus(status string) time.Duration {
	switch status {
	case spec.BOOKED:
		return s.config.IntervalBooked
	case spec.DEPARTED:
		return s.config.IntervalDeparted
	case spec.IN_TRANSIT:
		return s.config.IntervalInTransit
	case spec.ARRIVED:
		return s.config.IntervalArrived
	case spec.DELIVERED:
		return 0
	default:
		return s.config.IntervalInTransit
	}
}

func parseEnvInt(key string, fallback int) int {
	valStr := strings.TrimSpace(os.Getenv(key))
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val < 0 {
		return fallback
	}
	return val
}
