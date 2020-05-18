package gang

import (
	"context"
	"sync"
)

// Package gang provides utilities to manage multiple goroutines.
type Gang struct {
	ContinueOnErrExit   bool
	ContinueOnCleanExit bool

	fns     []func(context.Context) error
	fnMutex sync.Mutex

	errCh chan error
	once  sync.Once
}

// withLock runs fn under fnMutex lock.
func (g *Gang) withLock(fn func()) {
	g.fnMutex.Lock()
	defer g.fnMutex.Unlock()
	fn()
}

// AddWithChan adds fn to Gang. fn will be run when Run() is invoked.
func (g *Gang) AddWithChan(fn func(<-chan struct{})) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			fn(ctx.Done())
			return nil
		})
	})
}

// AddWithCtx adds fn to Gang. fn will be run when Run() is invoked.
func (g *Gang) AddWithCtx(fn func(context.Context)) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			fn(ctx)
			return nil
		})
	})
}

// AddWithChanE adds fn to Gang. fn will be run when Run() is invoked.
func (g *Gang) AddWithChanE(fn func(<-chan struct{}) error) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			return fn(ctx.Done())
		})
	})
}

// AddWithCtxE adds fn to Gang. fn will be run when Run() is invoked.
func (g *Gang) AddWithCtxE(fn func(context.Context) error) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			return fn(ctx)
		})
	})
}

// Run executes all functions that have been added to the
// Gang via Add* functions. Each function executes in it's own goroutines.
// If the provided context is cancelled, context (or done) channel passed to
// added function is cancelled (or closed) to cleanly terminate all functions.
// If any of the function errors out or exits, then the context
// (or channel) passed to added function is cancelled (or closed) as well.
// The returned channel returns errors (if any) returned by all of
// added functions and is closed once all the functions have finished executing.
func (g *Gang) Run(ctx context.Context) <-chan error {
	g.once.Do(func() {
		var wg sync.WaitGroup
		g.errCh = make(chan error, len(g.fns))
		ctx, cancel := context.WithCancel(ctx)
		wg.Add(len(g.fns))

		go func(errCh chan error) {
			wg.Wait()
			close(errCh)
		}(g.errCh)

		for _, fn := range g.fns {
			go func(fn func(ctx context.Context) error) {
				defer wg.Done()

				err := fn(ctx)

				if err != nil {
					if !g.ContinueOnErrExit {
						cancel()
					}
					g.errCh <- err
				} else {
					if !g.ContinueOnCleanExit {
						cancel()
					}
				}
			}(fn)
		}
	})
	return g.errCh
}
