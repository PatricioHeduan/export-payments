package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

// JobFunc defines the function signature executed by the scheduler.
type JobFunc func(ctx context.Context) error

// Scheduler coordinates timed background execution of tasks.
type Scheduler struct {
	cron    *cron.Cron
	jobFunc JobFunc
	cronExp string
	mu      sync.Mutex
	running bool
}

// NewScheduler creates a scheduler configured with the specified cron expression and task function.
// It wraps execution with recovery and skip-if-running middleware to prevent overlapping executions.
func NewScheduler(cronExp string, job JobFunc) *Scheduler {
	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))

	return &Scheduler{
		cron:    c,
		jobFunc: job,
		cronExp: cronExp,
	}
}

// Start registers the scheduled job and starts the cron engine.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.cron.AddFunc(s.cronExp, func() {
		log.Println("[Scheduler] Scheduled execution triggered.")
		if err := s.jobFunc(ctx); err != nil {
			log.Printf("[Scheduler] Error during scheduled execution: %v", err)
			return
		}
		log.Println("[Scheduler] Scheduled execution completed successfully.")
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.running = true
	log.Printf("[Scheduler] Scheduler started successfully with schedule expression: '%s'", s.cronExp)
	return nil
}

// Stop halts the scheduler and returns a context that resolves when running jobs complete.
func (s *Scheduler) Stop() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}

	log.Println("[Scheduler] Stopping scheduler...")
	stopCtx := s.cron.Stop()
	s.running = false
	return stopCtx
}
