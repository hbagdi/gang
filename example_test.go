package gang_test

import (
	"context"
	"fmt"

	"github.com/hbagdi/gang"
)

func ExampleGang_Run() {
	var g gang.Gang

	g.AddWithCtxE(func(ctx context.Context) error {
		fmt.Println("foo")
		return nil
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		fmt.Println("bar")
		return nil
	})

	errCh := g.Run(context.Background())
	<-errCh
}

func ExampleGang_Run_errCh() {
	var g gang.Gang

	g.AddWithCtxE(func(ctx context.Context) error {
		fmt.Println("foo")
		return fmt.Errorf("err1")
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		fmt.Println("bar")
		return fmt.Errorf("err2")
	})

	errCh := g.Run(context.Background())
	for err := range errCh {
		fmt.Println(err)
	}
}
