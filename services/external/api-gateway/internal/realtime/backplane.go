package realtime

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// Config defines realtime backplane behavior.
type Config struct {
	DSN             string
	Channel         string
	CleanupInterval time.Duration
	Retention       time.Duration
}

// Backplane consumes PostgreSQL LISTEN/NOTIFY and fanouts events to local subscribers.
type Backplane struct {
	cfg    Config
	repo   *Repository
	logger *slog.Logger

	mu         sync.RWMutex
	nextSubID  int64
	subs       map[int64]chan Event
	running    bool
	cancelFunc context.CancelFunc
}

// NewBackplane constructs realtime backplane.
func NewBackplane(cfg Config, repo *Repository, logger *slog.Logger) *Backplane {
	channel := strings.TrimSpace(cfg.Channel)
	if channel == "" {
		channel = "codex_realtime"
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 10 * time.Minute
	}
	if cfg.Retention <= 0 {
		cfg.Retention = 72 * time.Hour
	}
	cfg.Channel = channel

	if logger == nil {
		logger = slog.Default()
	}
	return &Backplane{
		cfg:    cfg,
		repo:   repo,
		logger: logger,
		subs:   make(map[int64]chan Event),
	}
}

// Start launches listener and cleanup loops.
func (b *Backplane) Start(ctx context.Context) {
	if b == nil {
		return
	}

	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	b.cancelFunc = cancel
	b.running = true
	b.mu.Unlock()

	go b.listenLoop(runCtx)
	go b.cleanupLoop(runCtx)
}

// Stop gracefully stops background loops.
func (b *Backplane) Stop() {
	if b == nil {
		return
	}
	b.mu.Lock()
	cancel := b.cancelFunc
	b.cancelFunc = nil
	b.running = false
	for id, ch := range b.subs {
		delete(b.subs, id)
		close(ch)
	}
	b.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Subscribe registers one subscriber channel.
func (b *Backplane) Subscribe(buffer int) (<-chan Event, func()) {
	if b == nil {
		ch := make(chan Event)
		close(ch)
		return ch, func() {}
	}
	if buffer <= 0 {
		buffer = 64
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextSubID++
	id := b.nextSubID
	ch := make(chan Event, buffer)
	b.subs[id] = ch

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		existing, ok := b.subs[id]
		if !ok {
			return
		}
		delete(b.subs, id)
		close(existing)
	}
	return ch, unsubscribe
}

func (b *Backplane) emit(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- event:
		default:
			// Slow consumer: drop event to avoid global backpressure.
		}
	}
}

func (b *Backplane) listenLoop(ctx context.Context) {
	retryDelay := time.Second
	maxRetryDelay := 10 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := pgx.Connect(ctx, b.cfg.DSN)
		if err != nil {
			b.logger.Error("realtime LISTEN connection failed", "err", err)
			if !sleepOrDone(ctx, retryDelay) {
				return
			}
			retryDelay = minDuration(retryDelay*2, maxRetryDelay)
			continue
		}

		if _, err := conn.Exec(ctx, "LISTEN "+b.cfg.Channel); err != nil {
			b.logger.Error("realtime LISTEN subscribe failed", "channel", b.cfg.Channel, "err", err)
			_ = conn.Close(ctx)
			if !sleepOrDone(ctx, retryDelay) {
				return
			}
			retryDelay = minDuration(retryDelay*2, maxRetryDelay)
			continue
		}

		retryDelay = time.Second
		b.logger.Info("realtime LISTEN started", "channel", b.cfg.Channel)

		for {
			if ctx.Err() != nil {
				_ = conn.Close(ctx)
				return
			}

			notification, waitErr := conn.WaitForNotification(ctx)
			if waitErr != nil {
				b.logger.Warn("realtime LISTEN interrupted", "err", waitErr)
				_ = conn.Close(ctx)
				break
			}
			if notification == nil {
				continue
			}
			eventID, convErr := strconv.ParseInt(strings.TrimSpace(notification.Payload), 10, 64)
			if convErr != nil || eventID <= 0 {
				b.logger.Warn("realtime notify payload is invalid", "payload", notification.Payload, "err", convErr)
				continue
			}
			event, ok, getErr := b.repo.GetByID(ctx, eventID)
			if getErr != nil {
				b.logger.Warn("realtime fetch event failed", "event_id", eventID, "err", getErr)
				continue
			}
			if !ok {
				continue
			}
			b.emit(event)
		}
	}
}

func (b *Backplane) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(b.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().UTC().Add(-b.cfg.Retention)
			deleted, err := b.repo.CleanupOlderThan(ctx, cutoff)
			if err != nil {
				b.logger.Warn("realtime cleanup failed", "err", err)
				continue
			}
			if deleted > 0 {
				b.logger.Info("realtime cleanup completed", "deleted", deleted, "cutoff", cutoff.Format(time.RFC3339))
			}
		}
	}
}

func sleepOrDone(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
