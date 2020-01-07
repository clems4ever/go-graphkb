package utils

import "sync"

type SynchronizedMap struct {
	m    map[string]interface{}
	lock sync.Mutex
}

func NewSynchronizedMap() *SynchronizedMap {
	return &SynchronizedMap{
		m: make(map[string]interface{}),
	}
}

func (sm *SynchronizedMap) Set(key string, value interface{}) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.m[key] = value
}

func (sm *SynchronizedMap) Get(key string) (interface{}, bool) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	v, ok := sm.m[key]
	return v, ok
}
