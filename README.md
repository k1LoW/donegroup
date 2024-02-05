# donegroup [![Go Reference](https://pkg.go.dev/badge/github.com/k1LoW/donegroup.svg)](https://pkg.go.dev/github.com/k1LoW/donegroup) [![CI](https://github.com/k1LoW/donegroup/actions/workflows/ci.yml/badge.svg)](https://github.com/k1LoW/donegroup/actions/workflows/ci.yml) ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/time.svg)

`donegroup` is a package that provides a graceful cleanup transaction to context.Context when the context is canceled ( **done** ).

> errgroup.Group after <-ctx.Done() = donegroup

## Usage

Use [donegroup.WithCancel](https://pkg.go.dev/github.com/k1LoW/donegroup#WithCancel) instead of [context.WithCancel](https://pkg.go.dev/context#WithCancel).

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/k1LoW/donegroup"
)

func main() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		// Cleanup process of some kind
		fmt.Println("cleanup 1")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		// Cleanup process of some kind
		time.Sleep(1 * time.Second)
		fmt.Println("cleanup 2")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Main process of some kind

	defer func() {
		cancel()
		if err := donegroup.Wait(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("finish")

	// Output:
	// finish
	// cleanup 1
	// cleanup 2
}
```

[dongroup.Cleanup](https://pkg.go.dev/github.com/k1LoW/donegroup#Cleanup) is similar in usage to [testing.(*T) Cleanup](https://pkg.go.dev/testing#T.Cleanup), but the order of execution is not guaranteed.
