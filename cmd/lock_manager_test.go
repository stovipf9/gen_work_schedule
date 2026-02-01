package cmd

import (
	"sync"
	"testing"
)

func TestLockManager_GetLock(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		key             string
		getReservedLock func(*LockManager[string]) *sync.RWMutex
		assert          func(locks ...*sync.RWMutex)
	}{
		{
			name:            "空でないロックが返る",
			key:             "key",
			getReservedLock: func(*LockManager[string]) *sync.RWMutex { return nil },
			assert: func(locks ...*sync.RWMutex) {
				if locks[0] == nil {
					t.Errorf("could not get non-nil lock")
				}
			},
		},
		{
			name: "同一キーに対し既存のロックが返る",
			key:  "key",
			getReservedLock: func(lm *LockManager[string]) *sync.RWMutex {
				return lm.GetLock("key")
			},
			assert: func(locks ...*sync.RWMutex) {
				if locks[0] != locks[1] {
					t.Errorf("could not the reserved lock for the same key")
				}
			},
		},
		{
			name: "異なるキーに対し異なるロックが返る",
			key:  "another_key",
			getReservedLock: func(lm *LockManager[string]) *sync.RWMutex {
				return lm.GetLock("key")
			},
			assert: func(locks ...*sync.RWMutex) {
				if locks[0] != nil && locks[0] == locks[1] {
					t.Errorf("errorneously returned the same lock for different keys")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLockManager(func(key string) string { return key })
			reservedLock := tt.getReservedLock(lm)
			got := lm.GetLock(tt.key)
			tt.assert(got, reservedLock)
		})
	}
}
