package cmd

import "sync"

type LockManager[T any] struct {
	locks  map[string]*sync.RWMutex // 読み書きロック
	mu     sync.Mutex               // 読み書きロック自体のロック
	genKey func(T) string
}

func NewLockManager[T any](genkey func(T) string) *LockManager[T] {
	return &LockManager[T]{
		locks:  make(map[string]*sync.RWMutex),
		genKey: genkey,
	}
}

func (lm *LockManager[T]) GetLock(key T) *sync.RWMutex {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	keyStr := lm.genKey(key)

	if lock, exists := lm.locks[keyStr]; exists {
		// 既存の読み書きロック
		return lock
	}
	// 新規の読み書きロック
	lock := &sync.RWMutex{}
	lm.locks[keyStr] = lock
	return lock
}
