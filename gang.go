package gang

import (
	"context"
	"sync"
)

type Gang struct {
	ContinueOnErrExit   bool
	ContinueOnCleanExit bool

	fns     []func(context.Context) error
	fnMutex sync.Mutex

	errCh chan error
	once  sync.Once
}

func (g *Gang) withLock(fn func()) {
	g.fnMutex.Lock()
	defer g.fnMutex.Unlock()
	fn()
}

func (g *Gang) AddWithChan(fn func(<-chan struct{})) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			fn(ctx.Done())
			return nil
		})
	})
}

func (g *Gang) AddWithCtx(fn func(context.Context)) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			fn(ctx)
			return nil
		})
	})
}

func (g *Gang) AddWithChanE(fn func(<-chan struct{}) error) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			return fn(ctx.Done())
		})
	})
}

func (g *Gang) AddWithCtxE(fn func(context.Context) error) {
	g.withLock(func() {
		g.fns = append(g.fns, func(ctx context.Context) error {
			return fn(ctx)
		})
	})
}

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
