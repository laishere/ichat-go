package sched

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"ichat-go/di"
	"ichat-go/logging"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type RLock interface {
	Lock() bool
	Unlock() bool
}

type lock struct {
	key       string
	id        uint64
	c         *redis.Client
	heartbeat time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
	logger    logging.Logger
	mu        sync.Mutex
	locked    atomic.Bool
}

func NewLock(key string, heartbeat time.Duration) RLock {
	c := di.ENV().RDB()
	l := &lock{key: "lock:" + key, c: c, heartbeat: heartbeat}
	l.logger = logging.NewLogger(l.key)
	return l
}

func (l *lock) Lock() bool {
	l.mu.Lock()
	l.id = rand.Uint64()
	l.ctx, l.cancel = context.WithCancel(context.Background())
	if !l.tryLock() {
		l.mu.Unlock()
		return false
	}
	go l.heartbeatLoop()
	l.locked.Store(true)
	return true
}

func (l *lock) ttl() time.Duration {
	return l.heartbeat * 2
}

func (l *lock) tryLock() bool {
	t := time.NewTicker(time.Millisecond * 100)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			r, err := l.c.SetNX(l.ctx, l.key, l.id, l.ttl()).Result()
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					l.logger.Error("Failed to lock", err)
				}
				return false
			}
			if r {
				return true
			}
		case <-l.ctx.Done():
			return false
		}
	}
}

func (l *lock) heartbeatLoop() {
	t := time.NewTicker(l.heartbeat)
	defer t.Stop()
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-t.C:
			l.c.Expire(l.ctx, l.key, l.ttl())
		}
	}
}

func (l *lock) checkAndDelete() bool {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		end
		return 0
	`
	r, err := l.c.Eval(context.Background(), script, []string{l.key}, l.id).Result()
	if err != nil {
		return false
	}
	return r.(int64) == 1
}

func (l *lock) Unlock() bool {
	if l.cancel != nil {
		l.cancel()
	}
	r := l.checkAndDelete()
	if l.locked.Load() {
		l.mu.Unlock()
		l.locked.Store(false)
	}
	return r
}
