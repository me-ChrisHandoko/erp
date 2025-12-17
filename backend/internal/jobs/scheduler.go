package jobs

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"backend/internal/config"
)

// Scheduler manages background jobs using cron
type Scheduler struct {
	cron        *cron.Cron
	db          *gorm.DB
	config      *config.Config
	lastCleanup time.Time
	isRunning   bool
}

// NewScheduler creates a new job scheduler
func NewScheduler(db *gorm.DB, cfg *config.Config) *Scheduler {
	return &Scheduler{
		db:     db,
		config: cfg,
		// Use UTC timezone and enable seconds field (6-field cron format)
		cron: cron.New(
			cron.WithLocation(time.UTC),
			cron.WithSeconds(), // Enable 6-field format: second minute hour day month weekday
		),
	}
}

// Start starts the scheduler and registers all jobs
func (s *Scheduler) Start() error {
	if !s.config.Job.EnableCleanup {
		log.Println("[JOB] Background jobs disabled via config")
		return nil
	}

	log.Println("[JOB] Starting background job scheduler...")

	// Register cleanup jobs
	if _, err := s.cron.AddFunc(s.config.Job.RefreshTokenCleanup, s.cleanupExpiredRefreshTokens); err != nil {
		return err
	}

	if _, err := s.cron.AddFunc(s.config.Job.EmailCleanup, s.cleanupExpiredEmailVerifications); err != nil {
		return err
	}

	if _, err := s.cron.AddFunc(s.config.Job.PasswordCleanup, s.cleanupExpiredPasswordResets); err != nil {
		return err
	}

	if _, err := s.cron.AddFunc(s.config.Job.LoginCleanup, s.cleanupOldLoginAttempts); err != nil {
		return err
	}

	// Start the scheduler
	s.cron.Start()
	s.isRunning = true

	log.Printf("[JOB] Scheduler started with %d jobs", len(s.cron.Entries()))
	log.Printf("[JOB] Refresh token cleanup: %s", s.config.Job.RefreshTokenCleanup)
	log.Printf("[JOB] Email cleanup: %s", s.config.Job.EmailCleanup)
	log.Printf("[JOB] Password cleanup: %s", s.config.Job.PasswordCleanup)
	log.Printf("[JOB] Login cleanup: %s", s.config.Job.LoginCleanup)

	return nil
}

// Stop stops the scheduler gracefully
// Returns a context that is done when all running jobs have completed
func (s *Scheduler) Stop() context.Context {
	log.Println("[JOB] Stopping background job scheduler...")
	ctx := s.cron.Stop()
	s.isRunning = false
	log.Println("[JOB] Scheduler stopped")
	return ctx
}

// IsRunning returns true if the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	return s.isRunning
}

// GetLastCleanupTime returns the timestamp of the last cleanup execution
func (s *Scheduler) GetLastCleanupTime() time.Time {
	return s.lastCleanup
}
