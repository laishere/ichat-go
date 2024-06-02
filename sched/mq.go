package sched

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"ichat-go/di"
	"ichat-go/logging"
	"sync"
	"time"
)

type MQ interface {
	SaveState(state int)
	State() int
	Push(message Message) error
	PushIfStateExits(message Message) error
	Ack(success bool)
	Channel() <-chan Message
	Close(clear bool)
	Expire(d time.Duration)
	ClearExpire()
}

type mq struct {
	c        *redis.Client
	key      string
	stateKey string
	mu       sync.Mutex
	lck      RLock
	ch       chan Message
	ctx      context.Context
	cancel   context.CancelFunc
	logger   logging.Logger
}

func NewMQ(key string) MQ {
	m := &mq{c: di.ENV().RDB(), key: "mq:" + key}
	m.stateKey = m.key + ":state"
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.logger = logging.NewLogger(m.key)
	m.lck = NewLock(m.key, time.Second*30)
	return m
}

func (m *mq) SaveState(state int) {
	m.c.Set(m.ctx, m.stateKey, state, 0)
}

func (m *mq) State() int {
	r, err := m.c.Get(m.ctx, m.stateKey).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0
		}
		m.logger.Error("Failed to get state", err)
	}
	return r
}

func (m *mq) Push(message Message) error {
	_, err := m.c.RPush(m.ctx, m.key, toJson(message)).Result()
	if err != nil && !errors.Is(err, context.Canceled) {
		m.logger.Error("Failed to push message", err)
	}
	return err
}

func (m *mq) PushIfStateExits(message Message) error {
	ttl := m.c.TTL(m.ctx, m.stateKey).Val()
	if ttl == -2 {
		return errors.New("state key not exits")
	}
	err := m.Push(message)
	if err == nil && ttl != -1 {
		// 同步队列过期时间
		m.c.Expire(m.ctx, m.key, ttl)
	}
	return err
}

func (m *mq) first() (Message, error) {
	if !m.lck.Lock() {
		select {
		case <-m.ctx.Done():
			return Message{}, context.Canceled
		default:
			return Message{}, redis.Nil
		}
	}
	t := time.NewTicker(time.Millisecond * 100)
	defer t.Stop()
	for {
		r, err := m.c.LIndex(m.ctx, m.key, 0).Result()
		if err == nil {
			return fromJson(r), nil
		}
		select {
		case <-t.C:
			continue
		case <-m.ctx.Done():
			return Message{}, context.Canceled
		}
	}
}

func (m *mq) subscribe() {
	defer close(m.ch)
	for {
		r, err := m.first()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if !errors.Is(err, redis.Nil) {
				m.logger.Error("Unknown error", err)
			}
		}
		m.ch <- r
	}
}

func (m *mq) Ack(success bool) {
	if success {
		_, err := m.c.LPop(m.ctx, m.key).Result()
		if err != nil {
			m.logger.Error("Failed to pop acked message", err)
		}
	}
	m.lck.Unlock()
}

func (m *mq) Channel() <-chan Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ch == nil {
		m.ch = make(chan Message)
		go m.subscribe()
	}
	return m.ch
}

func (m *mq) Close(clear bool) {
	m.cancel()
	m.lck.Unlock()
	if clear {
		m.c.Del(context.Background(), m.key, m.stateKey) // no cancel
	}
}

func (m *mq) Expire(d time.Duration) {
	c := context.Background()
	m.c.Expire(c, m.key, d)
	m.c.Expire(c, m.stateKey, d)
}

func (m *mq) ClearExpire() {
	c := context.Background()
	m.c.Persist(c, m.key)
	m.c.Persist(c, m.stateKey)
}
