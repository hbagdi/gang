# gang

[![Build Status](https://github.com/hbagdi/gang/workflows/Test/badge.svg)](https://github.com/hbagdi/gang/actions?query=branch%3Amaster+event%3Apush)
[![GoDoc](https://godoc.org/github.com/hbagdi/gang?status.svg)](https://godoc.org/github.com/hbagdi/gang)

Package gang provides utilities to manage multiple goroutines.

```go
package main

import (
	"context"
	"fmt"

	"github.com/hbagdi/gang"
)

func main() {
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
```
