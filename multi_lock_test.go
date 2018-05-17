package main

import (
	"runtime"
	"sync"

	"gopkg.in/check.v1"
)

func (*S) TestMultiLocker(c *check.C) {
	locker := mLocker{m: make(map[string]*sync.Mutex)}
	locker.Lock("user@tsuru.io")
	locker.Lock("another.user@tsuru.io")
	locker.Unlock("user@tsuru.io")
	locker.Unlock("another.user@tsuru.io")
}

func (*S) TestMultiLockerSingle(c *check.C) {
	locker := mLocker{m: make(map[string]*sync.Mutex)}
	locker.Lock("user@tsuru.io")
	locker.Unlock("user@tsuru.io")
	locker.Lock("user@tsuru.io")
	locker.Unlock("user@tsuru.io")
}

func (*S) TestMultiLockerUsage(c *check.C) {
	locker := mLocker{m: make(map[string]*sync.Mutex)}
	locker.Lock("user@tsuru.io")
	count := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		locker.Lock("user@tsuru.io")
		count--
		locker.Unlock("user@tsuru.io")
		wg.Done()
	}()
	runtime.Gosched()
	c.Assert(count, check.Equals, 0)
	locker.Unlock("user@tsuru.io")
	wg.Wait()
	c.Assert(count, check.Equals, -1)
}

func (*S) TestMultiLockerUnlocked(c *check.C) {
	defer func() {
		r := recover()
		c.Assert(r, check.NotNil)
	}()
	locker := mLocker{m: make(map[string]*sync.Mutex)}
	locker.Unlock("user@tsuru.io")
}

func (*S) TestMultiLockerFunction(c *check.C) {
	locker := multiLocker()
	locker.Lock("user@tsuru.io")
	defer locker.Unlock("user@tsuru.io")
}
