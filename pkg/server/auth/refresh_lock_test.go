package auth

import (
	"sync"
	"sync/atomic"
	"testing"
)

type refreshLocks struct {
	refreshing sync.Mutex
	refreshes  map[string]sync.Mutex
}

func (r *refreshLocks) tryLock(name string) bool {
	r.refreshing.Lock()
	defer r.refreshing.Unlock()

	namedLock := r.refreshes[name]

	locked := namedLock.TryLock()
	r.refreshes[name] = namedLock

	return locked
}

func (r *refreshLocks) unlock(name string) {
	r.refreshing.Lock()
	defer r.refreshing.Unlock()

	namedLock, ok := r.refreshes[name]
	if !ok {
		return
	}

	namedLock.Unlock()
	delete(r.refreshes, name)
}

func newRefreshLocker() *refreshLocks {
	return &refreshLocks{
		refreshes: map[string]sync.Mutex{},
	}
}

func TestRefreshLocks(t *testing.T) {
	locks := newRefreshLocker()

	var wg sync.WaitGroup
	wg.Add(3)

	var refreshCount int32

	names := []string{"testing1", "testing2", "testing1"}

	for i := 0; i < 3; i++ {
		go func(v int) {
			defer wg.Done()

			if !locks.tryLock(names[v]) {
				return
			}
			atomic.AddInt32(&refreshCount, 1)
			t.Logf("unlocking %s", names[v])
			locks.unlock(names[v])
		}(i)
	}

	wg.Wait()

	if refreshCount != 2 {
		t.Fatalf("refreshCount got %d, want %d", refreshCount, 2)
	}
}
