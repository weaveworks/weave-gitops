package auth

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestRefreshLocks_Lock(t *testing.T) {
	locks := newRefreshLocker()

	var wg sync.WaitGroup
	wg.Add(3)

	var refreshCount int32

	names := []string{"testing1", "testing2", "testing1"}

	for i := 0; i < 3; i++ {
		go func(v int) {
			defer wg.Done()

			unlock := locks.lock(names[v])
			defer func() {
				t.Logf("unlocking %s", names[v])
				unlock()
			}()

			atomic.AddInt32(&refreshCount, 1)
		}(i)
	}

	wg.Wait()

	if refreshCount != 3 {
		t.Fatalf("refreshCount got %d, want %d", refreshCount, 3)
	}
}
