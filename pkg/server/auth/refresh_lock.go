package auth

import (
	"sync"
)

type refreshLocks struct {
	mutexes sync.Map
}

func (r *refreshLocks) lock(name string) func() {
	value, _ := r.mutexes.LoadOrStore(name, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}

func newRefreshLocker() *refreshLocks {
	return &refreshLocks{}
}
