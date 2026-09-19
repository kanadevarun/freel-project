package automations

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Scheduler interface {
	Start()
	Stop()
}

type scheduler struct {
	repo     Repository
	svc      Service
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

func NewScheduler(repo Repository, svc Service, interval time.Duration) Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &scheduler{
		repo:     repo,
		svc:      svc,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	log.Printf("🕒 [Automation Scheduler] Initialized. Polling interval: %v", s.interval)

	s.wg.Add(1)
	go s.run()
}

func (s *scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	log.Println("🛑 [Automation Scheduler] Stopped gracefully.")
}

func (s *scheduler) run() {
	defer s.wg.Done()

	// Initial delay on startup
	select {
	case <-time.After(5 * time.Second):
	case <-s.stopCh:
		return
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial check
	s.checkAndExecute()

	for {
		select {
		case <-ticker.C:
			s.checkAndExecute()
		case <-s.stopCh:
			return
		}
	}
}

func (s *scheduler) checkAndExecute() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dueAutos, err := s.repo.GetDueAutomations(ctx, 10)
	if err != nil {
		log.Printf("⚠️ [Automation Scheduler] Error polling due automations: %v", err)
		return
	}

	if len(dueAutos) == 0 {
		return
	}

	log.Printf("🔍 [Automation Scheduler] Found %d due automations to trigger", len(dueAutos))

	for _, auto := range dueAutos {
		s.dispatchSingleAutomation(ctx, auto)
	}
}

func (s *scheduler) dispatchSingleAutomation(ctx context.Context, auto *Automation) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🔥 [Automation Scheduler] Recovered from panic executing automation #%d: %v", auto.ID, r)
		}
	}()

	// Double check no active run exists
	active, err := s.repo.GetActiveExecutionForAutomation(ctx, auto.OrgID, auto.ID)
	if err != nil || active != nil {
		return
	}

	corrID := fmt.Sprintf("exec-sched-%d-%d", auto.ID, time.Now().UnixNano())
	exec := &AutomationExecution{
		AutomationID:      auto.ID,
		OrgID:             auto.OrgID,
		CorrelationID:     corrID,
		TriggerType:       TriggerTypeScheduled,
		TriggeredByUserID: nil,
		Status:            ExecutionStatusQueued,
	}

	createdExec, err := s.repo.CreateExecution(ctx, exec)
	if err != nil {
		log.Printf("⚠️ [Automation Scheduler] Failed to enqueue execution for #%d: %v", auto.ID, err)
		return
	}

	log.Printf("🚀 [Automation Scheduler] Enqueued scheduled execution #%d for '%s' (#%d)",
		createdExec.ID, auto.Name, auto.ID)

	// Execute job
	if err := s.svc.ExecuteJob(ctx, auto, createdExec); err != nil {
		log.Printf("❌ [Automation Scheduler] Job execution error for #%d: %v", auto.ID, err)
	}
}
