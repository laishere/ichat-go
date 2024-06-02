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

type DQ interface {
	Schedule(time time.Time, message Message) error
	Delay(duration time.Duration, message Message) error
	Delete(id string)
	Channel() <-chan DelayMessage
	Close(clear bool)
}

type dq struct {
	c      *redis.Client
	zKey   string
	hKey   string
	subKey string
	ch     chan DelayMessage
	cancel func()
	mu     sync.Mutex
	logger logging.Logger
}

func NewDQ(key string) DQ {
	ctx, cancel := context.WithCancel(context.Background())
	key = "dq:" + key
	logger := logging.NewLogger(key)
	return &dq{c: di.ENV().RDB().WithContext(ctx), zKey: key + ":z", hKey: key + ":h", subKey: key + ":sub",
		cancel: cancel, logger: logger}
}

func (d *dq) handleErr(msg string, err error) {
	if !errors.Is(err, redis.Nil) && !errors.Is(err, context.Canceled) {
		d.logger.Error(msg, err.Error())
	}
}

func (d *dq) Schedule(time time.Time, message Message) error {
	if message.Id == "" {
		panic("dq Schedule: message.Id is not set")
	}
	//d.logger.Debug("Schedule ", time.String())
	_, err := d.c.TxPipelined(d.c.Context(), func(pipe redis.Pipeliner) error {
		pipe.HSet(d.c.Context(), d.hKey, message.Id, toJson(message))
		pipe.ZAdd(d.c.Context(), d.zKey, &redis.Z{Score: float64(time.UnixMilli()), Member: message.Id})
		pipe.Publish(d.c.Context(), d.subKey, "")
		return nil
	})
	if err != nil {
		d.handleErr("Schedule:", err)
	}
	return err
}

func (d *dq) Delay(duration time.Duration, message Message) error {
	return d.Schedule(time.Now().Add(duration), message)
}

func (d *dq) Delete(id string) {
	d.c.ZRem(d.c.Context(), d.zKey, id)
	d.c.HDel(d.c.Context(), d.hKey, id)
}

func (d *dq) pop(id string, t time.Time) {
	v, err := d.c.HGet(d.c.Context(), d.hKey, id).Result()
	if err != nil {
		d.handleErr("pop:", err)
		return
	}
	m := DelayMessage{fromJson(v), t}
	d.ch <- m
	d.Delete(id)
}

func (d *dq) poll() {
	delay := 0 * time.Second
	latency := time.Millisecond * 3
	sub := d.c.Subscribe(d.c.Context(), d.subKey)
	defer func() {
		close(d.ch)
		_ = sub.Close()
	}()
	for {
		waitCtx, cancel := context.WithTimeout(context.Background(), delay-latency)
		select {
		case <-d.c.Context().Done():
			cancel()
			return
		case <-sub.Channel():
			cancel()
			delay = 0
		case <-waitCtx.Done():
			v, err := d.c.BZPopMin(d.c.Context(), time.Second, d.zKey).Result()
			if err != nil {
				d.handleErr("poll:", err)
				if errors.Is(err, redis.Nil) {
					delay = 0
				} else {
					delay = time.Second
				}
				continue
			}
			t := time.UnixMilli(int64(v.Score))
			delay = time.Until(t)
			if delay > latency {
				d.c.ZAdd(d.c.Context(), d.zKey, &v.Z)
				continue
			}
			latency = max(0, latency-delay)
			id := v.Member.(string)
			d.pop(id, t)
			delay = 0
		}
	}
}

func (d *dq) Channel() <-chan DelayMessage {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ch == nil {
		d.ch = make(chan DelayMessage)
		go d.poll()
	}
	return d.ch
}

func (d *dq) Close(clear bool) {
	d.cancel()
	if clear {
		d.c.Del(context.Background(), d.zKey, d.hKey, d.subKey)
	}
}
