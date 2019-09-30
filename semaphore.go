// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file at https://golang.org/LICENSE?m=text

// This package provides the channel-based semaphore implementation from Bryan C. Mills' 2018 Gophercon talk.
package semaphore

import (
	"context"
	"sync"
)

type token struct{}

type Semaphore struct {
	tokens chan token
	wg     sync.WaitGroup
}

// New returns a semphore with n tokens.
func New(n int) *Semaphore {
	return &Semaphore{tokens: make(chan token, n)}
}

func (s *Semaphore) startWithToken(f func()) {
	s.wg.Add(1)
	go func() {
		f()
		<-s.tokens
		s.wg.Done()
	}()
}

// Add calls f in a goroutine while holding a semaphore token.
// If ctx is done and no token is available, Add returns ctx.Err().
//
// f must not call Add or MustAdd, but may call TryAdd.
func (s *Semaphore) Add(ctx context.Context, f func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.tokens <- token{}:
	}
	s.startWithToken(f)
	return nil
}

// TryAdd is like Add, but does not block.
// The return value reports whether f was added.
func (s *Semaphore) TryAdd(f func()) bool {
	select {
	case s.tokens <- token{}:
	default:
		return false
	}
	s.startWithToken(f)
	return true
}

// MustAdd is like Add, but blocks indefinitely.
func (s *Semaphore) MustAdd(f func()) {
	s.tokens <- token{}
	s.startWithToken(f)
}

// Wait blocks until all tokens are released.
func (s *Semaphore) Wait() {
	for len(s.tokens) > 0 {
		s.wg.Wait()
	}
}
