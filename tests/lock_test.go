package tests

import (
	"ichat-go/db"
	"ichat-go/sched"
	"sync"
	"testing"
	"time"
)

func assertLockSuccess(t *testing.T, l sched.RLock) {
	if !l.Lock() {
		t.Errorf("lock failed")
	}
}

func TestUnlockAhead(t *testing.T) {
	db.InitForTest()
	lock := sched.NewLock("test", time.Minute)
	if lock.Unlock() {
		t.Errorf("unlock should fail")
	}
}

func TestLockMutexAndUnlock(t *testing.T) {
	db.InitForTest()
	const hb = time.Second
	lock1 := sched.NewLock("test", hb)
	lock2 := sched.NewLock("test", hb)

	var t1, t2 time.Time

	wg := sync.WaitGroup{}
	wg.Add(2)

	work1 := func() {
		assertLockSuccess(t, lock1)
		t.Logf("lock1 success")
		defer func() {
			if lock1.Unlock() {
				t.Logf("unlock1 success")
			} else {
				t.Errorf("unlock1 failed")
			}
			wg.Done()
		}()
		time.Sleep(time.Millisecond * 1500)
		t1 = time.Now()
	}

	work2 := func() {
		time.Sleep(time.Millisecond * 50)
		assertLockSuccess(t, lock2)
		t.Logf("lock2 success")
		defer func() {
			if lock2.Unlock() {
				t.Logf("unlock2 success")
			} else {
				t.Errorf("unlock2 failed")
			}
			wg.Done()
		}()
		if lock1.Unlock() {
			t.Errorf("unlock1 should fail")
		}
		t2 = time.Now()
	}

	go work1()
	go work2()

	wg.Wait()

	if t1.After(t2) {
		t.Errorf("locks should be in order")
	}
}

func TestLockReentrant(t *testing.T) {
	db.InitForTest()
	const hb = time.Second
	lock := sched.NewLock("test", hb)

	work := func() {
		assertLockSuccess(t, lock)
		t.Logf("Lock 1")
		defer func() {
			if lock.Unlock() {
				t.Logf("Unlock 2")
			} else {
				t.Errorf("Unlock failed")
			}
		}()
		go func() {
			time.Sleep(time.Millisecond * 500)
			lock.Unlock()
			t.Log("Unlock 1")
		}()
		t.Log("Try to lock 2")
		assertLockSuccess(t, lock)
		time.Sleep(time.Millisecond * 100)
		t.Log("Lock 2")
	}

	work()
}

func TestLockCancel(t *testing.T) {
	db.InitForTest()
	const hb = time.Second
	lock1 := sched.NewLock("test", hb)
	lock2 := sched.NewLock("test", hb)

	wg := sync.WaitGroup{}
	wg.Add(2)

	work1 := func() {
		defer wg.Done()
		assertLockSuccess(t, lock1)
		t.Logf("lock1 success")
		defer func() {
			if lock1.Unlock() {
				t.Logf("unlock1 success")
			} else {
				t.Errorf("unlock1 failed")
			}
		}()
		time.Sleep(time.Millisecond * 500)
	}

	work2 := func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 50)
		go func() {
			time.Sleep(time.Millisecond * 100)
			// cancel locking
			if lock2.Unlock() {
				t.Errorf("unlock2 should fail")
			}
		}()
		if lock2.Lock() {
			t.Errorf("lock2 should fail")
		} else {
			t.Log("lock2 cancelled")
		}
	}

	go work1()
	go work2()

	wg.Wait()
}

func TestLockReuse(t *testing.T) {
	db.InitForTest()
	const hb = time.Second
	lock := sched.NewLock("test", hb)

	work := func() {
		assertLockSuccess(t, lock)
		t.Logf("lock success")
		if lock.Unlock() {
			t.Log("unlock success")
		} else {
			t.Errorf("unlock failed")
		}
	}

	work()
	work()
}

func Test(t *testing.T) {
	mu := sync.Mutex{}
	mu.Lock()
	t.Log("Locked")
	go func() {
		time.Sleep(time.Second)
		mu.Unlock()
		t.Log("Unlock")
	}()
	t.Log("Try to lock 2")
	mu.Lock()
	time.Sleep(time.Millisecond * 100)
	t.Log("Locked 2")
	mu.Unlock()
	t.Log("Unlock 2")
}
