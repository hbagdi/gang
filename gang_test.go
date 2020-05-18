package gang

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"
	"testing"
)

func TestZeroValue(t *testing.T) {
	var g Gang
	errCh := g.Run(context.Background())
	_, ok := <-errCh
	if ok {
		t.Errorf("zero value has an open error channel")
	}
}

func TestAddFns(t *testing.T) {
	var g Gang
	g.AddWithCtx(func(ctx context.Context) {
		<-ctx.Done()
	})
	g.AddWithCtxE(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})
	g.AddWithChan(func(done <-chan struct{}) {
		<-done
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		<-done
		return nil
	})
	if len(g.fns) != 4 {
		t.Errorf("unexpected length of fns")
	}
}

func TestDefault(t *testing.T) {
	var g Gang
	var counter uint64
	g.AddWithCtx(func(ctx context.Context) {
		atomic.AddUint64(&counter, 1)
	})
	g.AddWithCtxE(func(ctx context.Context) error {
		atomic.AddUint64(&counter, 1)
		return nil
	})
	g.AddWithChan(func(done <-chan struct{}) {
		atomic.AddUint64(&counter, 1)
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		atomic.AddUint64(&counter, 1)
		return nil
	})
	errCh := g.Run(context.Background())
	_, ok := <-errCh
	if ok {
		t.Errorf("errCh is open, expected it to be closed")
	}
	if 4 != atomic.LoadUint64(&counter) {
		t.Errorf("unexpected counter value")
	}
}

func TestErrorPropagation(t *testing.T) {
	var g Gang
	g.AddWithCtxE(func(ctx context.Context) error {
		return fmt.Errorf("first")
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		return fmt.Errorf("second")
	})
	errCh := g.Run(context.Background())
	var errs []string
	for err := range errCh {
		errs = append(errs, err.Error())
	}
	sort.Strings(errs)
	if !reflect.DeepEqual(errs, []string{"first", "second"}) {
		t.Errorf("unexpected error values")
	}

	if len(errs) != 2 {
		t.Errorf("unexpected number of errors")
	}
}

func TestContextCancellation(t *testing.T) {
	var g Gang
	g.AddWithCtxE(func(ctx context.Context) error {
		return fmt.Errorf("first")
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		<-done
		return nil
	})
	errCh := g.Run(context.Background())
	var errs []string
	for err := range errCh {
		errs = append(errs, err.Error())
	}

	if len(errs) != 1 {
		t.Errorf("unexpected number of errors")
	}
}

func TestContinueOnErrExit(t *testing.T) {
	var g Gang
	g.ContinueOnErrExit = true
	var counter uint64
	wait := make(chan error)
	g.AddWithChan(func(done <-chan struct{}) {
		select {
		case <-wait:
		case <-done:
		}
		atomic.AddUint64(&counter, 1)
	})
	g.AddWithCtxE(func(ctx context.Context) error {
		return fmt.Errorf("error number 42")
	})
	errCh := g.Run(context.Background())
	<-errCh
	close(wait)
	<-errCh
	if 1 != atomic.LoadUint64(&counter) {
		t.Errorf("unexpected counter value")
	}
}

func TestContinueOnCleanExit(t *testing.T) {
	var g Gang
	g.ContinueOnCleanExit = true
	var counter uint64
	wait := make(chan error)
	g.AddWithChan(func(done <-chan struct{}) {
		select {
		case <-wait:
		case <-done:
		}
		atomic.AddUint64(&counter, 1)
	})
	g.AddWithCtxE(func(ctx context.Context) error {
		close(wait)
		return nil
	})
	errCh := g.Run(context.Background())
	<-wait
	_, ok := <-errCh
	if ok {
		t.Errorf("errCh is open, expected it to be closed")
	}
	if 1 != atomic.LoadUint64(&counter) {
		t.Errorf("unexpected counter value")
	}
}
