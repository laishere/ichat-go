package tests

import (
	"context"
	"errors"
	"ichat-go/db"
	"ichat-go/sched"
	"strconv"
	"sync"
	"testing"
)

func TestMQ(t *testing.T) {
	db.InitForTest()
	m := sched.NewMQ("test")
	n := 10
	for i := 0; i < n; i++ {
		_ = m.Push(sched.Message{Type: 1, Payload: []byte(strconv.Itoa(i))})
	}
	actual := make([]string, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		cnt := 0
		for message := range m.Channel() {
			actual = append(actual, string(message.Payload))
			t.Logf("got %s", string(message.Payload))
			m.Ack(true)
			if cnt++; cnt == n {
				break
			}
		}
		wg.Done()
	}()
	wg.Wait()
	m.Close(true)
	for i, a := range actual {
		if a != strconv.Itoa(i) {
			t.Errorf("expected %s, got %s", strconv.Itoa(i), a)
		}
	}
	err := m.Push(sched.Message{Type: 1, Payload: []byte("hi")})
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
