package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/leksa/datamapper-senyar/internal/service"
	"github.com/leksa/datamapper-senyar/internal/sse"
)

// Mode represents the scheduler operating mode
type Mode string

const (
	ModeIdle   Mode = "idle"   // Night hours: 22:00 - 06:00, interval 20 minutes
	ModeNormal Mode = "normal" // Day hours without disaster: interval 3 minutes
	ModeActive Mode = "active" // Active disaster: interval 30 seconds
)

// Config holds scheduler configuration
type Config struct {
	IdleInterval   time.Duration // Default: 20 minutes
	NormalInterval time.Duration // Default: 3 minutes
	ActiveInterval time.Duration // Default: 30 seconds
	IdleStartHour  int           // Default: 22 (10 PM)
	IdleEndHour    int           // Default: 6 (6 AM)
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() *Config {
	return &Config{
		IdleInterval:   20 * time.Minute,
		NormalInterval: 3 * time.Minute,
		ActiveInterval: 30 * time.Second,
		IdleStartHour:  22,
		IdleEndHour:    6,
	}
}

// Scheduler handles automatic sync scheduling
type Scheduler struct {
	config          *Config
	syncService     *service.SyncService
	feedSyncService *service.FeedSyncService
	sseHub          *sse.Hub

	currentMode   Mode
	manualMode    *Mode // Manual override mode
	isRunning     bool
	lastSync      time.Time
	lastFeedSync  time.Time
	syncCount     int
	feedSyncCount int

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewScheduler creates a new scheduler
func NewScheduler(
	config *Config,
	syncService *service.SyncService,
	feedSyncService *service.FeedSyncService,
	sseHub *sse.Hub,
) *Scheduler {
	if config == nil {
		config = DefaultConfig()
	}

	return &Scheduler{
		config:          config,
		syncService:     syncService,
		feedSyncService: feedSyncService,
		sseHub:          sseHub,
		currentMode:     ModeNormal,
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.isRunning = true
	s.mu.Unlock()

	log.Println("[Scheduler] Starting...")

	// Initial sync
	go s.runSyncCycle()

	// Main loop
	go s.run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	log.Println("[Scheduler] Stopping...")
	s.cancel()
	s.isRunning = false
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	for {
		// Determine current mode and interval
		mode := s.determineMode()
		interval := s.getIntervalForMode(mode)

		s.mu.Lock()
		s.currentMode = mode
		s.mu.Unlock()

		log.Printf("[Scheduler] Mode: %s, Next sync in: %v", mode, interval)

		select {
		case <-s.ctx.Done():
			log.Println("[Scheduler] Stopped")
			return
		case <-time.After(interval):
			s.runSyncCycle()
		}
	}
}

// determineMode determines the current operating mode
func (s *Scheduler) determineMode() Mode {
	s.mu.RLock()
	manualMode := s.manualMode
	s.mu.RUnlock()

	// Check for manual override
	if manualMode != nil {
		return *manualMode
	}

	hour := time.Now().Hour()

	// Check if in idle hours (night time)
	if hour >= s.config.IdleStartHour || hour < s.config.IdleEndHour {
		return ModeIdle
	}

	// TODO: Check for active disaster flag from database or config
	// For now, default to normal mode during day hours
	return ModeNormal
}

// getIntervalForMode returns the sync interval for a given mode
func (s *Scheduler) getIntervalForMode(mode Mode) time.Duration {
	switch mode {
	case ModeIdle:
		return s.config.IdleInterval
	case ModeActive:
		return s.config.ActiveInterval
	default:
		return s.config.NormalInterval
	}
}

// runSyncCycle runs a complete sync cycle
func (s *Scheduler) runSyncCycle() {
	log.Println("[Scheduler] Running sync cycle...")

	// Broadcast sync start
	if s.sseHub != nil {
		s.sseHub.Broadcast("sync_start", map[string]interface{}{
			"mode": s.currentMode,
		})
	}

	var wg sync.WaitGroup
	var poskoResult, feedResult interface{}
	var poskoErr, feedErr error

	// Sync posko data
	wg.Add(1)
	go func() {
		defer wg.Done()
		poskoResult, poskoErr = s.syncService.SyncAll()
		if poskoErr != nil {
			log.Printf("[Scheduler] Posko sync error: %v", poskoErr)
		} else {
			s.mu.Lock()
			s.lastSync = time.Now()
			s.syncCount++
			s.mu.Unlock()
			log.Println("[Scheduler] Posko sync completed")
		}
	}()

	// Sync feed data
	wg.Add(1)
	go func() {
		defer wg.Done()
		feedResult, feedErr = s.feedSyncService.SyncAll()
		if feedErr != nil {
			log.Printf("[Scheduler] Feed sync error: %v", feedErr)
		} else {
			s.mu.Lock()
			s.lastFeedSync = time.Now()
			s.feedSyncCount++
			s.mu.Unlock()
			log.Println("[Scheduler] Feed sync completed")
		}
	}()

	wg.Wait()

	// Broadcast sync complete
	if s.sseHub != nil {
		s.sseHub.Broadcast("sync_complete", map[string]interface{}{
			"mode":        s.currentMode,
			"posko":       poskoResult,
			"posko_error": errorToString(poskoErr),
			"feed":        feedResult,
			"feed_error":  errorToString(feedErr),
		})
	}

	log.Println("[Scheduler] Sync cycle completed")
}

// SetMode manually sets the scheduler mode
func (s *Scheduler) SetMode(mode Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manualMode = &mode
	log.Printf("[Scheduler] Manual mode set to: %s", mode)
}

// ClearManualMode clears the manual mode override
func (s *Scheduler) ClearManualMode() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manualMode = nil
	log.Println("[Scheduler] Manual mode cleared, returning to automatic")
}

// SetActiveDisaster sets the scheduler to active mode for disaster response
func (s *Scheduler) SetActiveDisaster() {
	s.SetMode(ModeActive)
}

// ClearActiveDisaster clears the active disaster mode
func (s *Scheduler) ClearActiveDisaster() {
	s.ClearManualMode()
}

// GetStatus returns the current scheduler status
func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"is_running":      s.isRunning,
		"current_mode":    s.currentMode,
		"manual_override": s.manualMode != nil,
		"sync_count":      s.syncCount,
		"feed_sync_count": s.feedSyncCount,
	}

	if !s.lastSync.IsZero() {
		status["last_sync"] = s.lastSync
	}
	if !s.lastFeedSync.IsZero() {
		status["last_feed_sync"] = s.lastFeedSync
	}
	if s.manualMode != nil {
		status["manual_mode"] = *s.manualMode
	}

	return status
}

// TriggerSync manually triggers a sync cycle
func (s *Scheduler) TriggerSync() {
	go s.runSyncCycle()
}

func errorToString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
