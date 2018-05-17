package main

import "sync"

type mLocker struct {
	m   map[string]*sync.Mutex
	mut sync.Mutex
}

func multiLocker() *mLocker {
	return &mLocker{m: make(map[string]*sync.Mutex)}
}

func (l *mLocker) Lock(name string) {
	l.mut.Lock()
	mutex, ok := l.m[name]
	if !ok {
		mutex = new(sync.Mutex)
		l.m[name] = mutex
	}
	l.mut.Unlock()
	mutex.Lock()
}

func (l *mLocker) Unlock(name string) {
	l.mut.Lock()
	mutex := l.m[name]
	l.mut.Unlock()
	mutex.Unlock()
}
