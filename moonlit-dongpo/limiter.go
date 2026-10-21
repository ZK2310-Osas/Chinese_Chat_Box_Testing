package main

import (
	"context"
	"sync"
	"time"
)

type sessionState struct {
	lastStart time.Time
	active    bool
	touched   time.Time
}
type limiter struct {
	mu       sync.Mutex
	sessions map[string]*sessionState
	slots    chan struct{}
	stop     chan struct{}
}

func newLimiter() *limiter {
	l := &limiter{sessions: map[string]*sessionState{}, slots: make(chan struct{}, 25), stop: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.mu.Lock()
				for key, s := range l.sessions {
					if !s.active && time.Since(s.touched) > 10*time.Minute {
						delete(l.sessions, key)
					}
				}
				l.mu.Unlock()
			case <-l.stop:
				return
			}
		}
	}()
	return l
}
func (l *limiter) close() { close(l.stop) }
func (l *limiter) start(id string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	s := l.sessions[id]
	if s == nil {
		s = &sessionState{}
		l.sessions[id] = s
	}
	if s.active {
		return 0, false
	}
	wait := 5*time.Second - time.Since(s.lastStart)
	if s.lastStart.IsZero() || wait < 0 {
		wait = 0
	}
	s.lastStart = time.Now().Add(wait)
	s.active = true
	s.touched = time.Now()
	return wait, true
}
func (l *limiter) finish(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if s := l.sessions[id]; s != nil {
		s.active = false
		s.touched = time.Now()
	}
}
func (l *limiter) acquire(ctx context.Context) bool {
	timer := time.NewTimer(25 * time.Second)
	defer timer.Stop()
	select {
	case l.slots <- struct{}{}:
		return true
	case <-timer.C:
		return false
	case <-ctx.Done():
		return false
	}
}
func (l *limiter) release() { <-l.slots }
