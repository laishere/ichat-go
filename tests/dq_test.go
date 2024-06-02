package tests

import (
	"ichat-go/db"
	"ichat-go/sched"
	"sync"
	"testing"
	"time"
)

type dqTestCase struct {
	delay time.Duration
	id    string
}

func TestDQ(t *testing.T) {
	db.InitForTest()
	d := sched.NewDQ("test")
	cases := []dqTestCase{
		{delay: time.Second, id: "1"},
		{delay: time.Second * 2, id: "2"},
		{delay: time.Millisecond * 100, id: "3"},
	}
	expected := []string{"3", "1", "2"}
	timeTolerance := time.Millisecond * 10
	firstTimeTolerance := timeTolerance * 10
	go func() {
		for _, c := range cases {
			time.Sleep(time.Millisecond * 100)
			_ = d.Delay(c.delay, sched.Message{Id: c.id, Type: 1, Payload: []byte(c.id)})
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	actual := make([]string, 0)
	go func() {
		cnt := 0
		for m := range d.Channel() {
			actual = append(actual, m.Id)
			delay := time.Until(m.Time)
			t.Logf("got %s, time: %v, delay: %v", m.Id, time.Now(), delay)
			if delay < 0 {
				delay = -delay
			}
			if (cnt == 0 && delay > firstTimeTolerance) || (cnt != 0 && delay > timeTolerance) {
				t.Errorf("expected delay less than %v, got %v", timeTolerance, delay)
			}
			if cnt++; cnt == 3 {
				break
			}
		}
		println("done")
		wg.Done()
	}()
	wg.Wait()
	for i, a := range actual {
		if a != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], a)
		}
	}
	d.Close(true)
}
