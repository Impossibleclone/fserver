package storage

import (
	"sync"
)

type FileLocker struct {
	mu    sync.Mutex
	locks map[string]*sync.RWMutex
}

func NewFileLocker() *FileLocker {
	return &FileLocker{
		locks: make(map[string]*sync.RWMutex),
	}
}

func (l *FileLocker) getLock(filename string) *sync.RWMutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lock, exists := l.locks[filename]; exists {
		return lock
	}
	lock := &sync.RWMutex{}
	l.locks[filename] = lock
	return lock
}

func (l *FileLocker) LockRead(filename string) {
	l.getLock(filename).RLock()
}

func (l *FileLocker) UnlockRead(filename string) {
	l.getLock(filename).RUnlock()
}

func (l *FileLocker) LockWrite(filename string) {
	l.getLock(filename).Lock()
}

func (l *FileLocker) UnlockWrite(filename string) {
	l.getLock(filename).Unlock()
}
