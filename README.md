# gang

[![Build Status](https://github.com/hbagdi/gang/workflows/Test/badge.svg)](https://github.com/hbagdi/gang/actions?query=branch%3Amaster+event%3Apush)
[![GoDoc](https://godoc.org/github.com/hbagdi/gang?status.svg)](https://godoc.org/github.com/hbagdi/gang)

Package gang provides utilities to manage multiple goroutines.

## Synopsis

```go
package main

import (
	"context"
	"fmt"

	"github.com/hbagdi/gang"
)

func main() {
    // zero value is usable out of the box
	var g gang.Gang

    // add multiple functions which should be executed concurrently
	g.AddWithCtxE(func(ctx context.Context) error {
		fmt.Println("foo")
		return fmt.Errorf("err1")
	})
	g.AddWithChanE(func(done <-chan struct{}) error {
		fmt.Println("bar")
		return fmt.Errorf("err2")
	})

    // start execution of all the functions
	errCh := g.Run(context.Background())

    // collect errors from all the functions
	for err := range errCh {
		fmt.Println(err)
	}
    // execution finishes when errCh is closed
}
```

## Changelog

Changelog can be found in the [CHANGELOG.md](CHANGELOG.md) file.

## License

gang is licensed with Apache License Version 2.0.
Please read the [LICENSE](LICENSE) file for more details.
