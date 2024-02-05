# donegroup [![Go Reference](https://pkg.go.dev/badge/github.com/k1LoW/donegroup.svg)](https://pkg.go.dev/github.com/k1LoW/donegroup) [![CI](https://github.com/k1LoW/donegroup/actions/workflows/ci.yml/badge.svg)](https://github.com/k1LoW/donegroup/actions/workflows/ci.yml) ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/donegroup/time.svg)

`donegroup` is a package that provides a graceful cleanup transaction to context.Context when the context is canceled ( **done** ).

> errgroup.Group after <-ctx.Done() = donegroup

## Usage

Use [donegroup.WithCancel](https://pkg.go.dev/github.com/k1LoW/donegroup#WithCancel) instead of [context.WithCancel](https://pkg.go.dev/context#WithCancel).

Then, it can wait for the cleanup processes associated with the context using [donegroup.Wait](https://pkg.go.dev/github.com/k1LoW/donegroup#Wait).

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

	// Cleanup process 1 of some kind
	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		fmt.Println("cleanup 1")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Cleanup process 2 of some kind
	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
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

### Wait for a specified duration

Using [donegroup.WaitWithTimeout](https://pkg.go.dev/github.com/k1LoW/donegroup#WaitWithTimeout), it is possible to set a timeout for the cleanup processes.

Note that each cleanup process must handle its own context argument.

```go
ctx, cancel := WithCancel(context.Background())

// Cleanup process of some kind
if err := Clenup(ctx, func(ctx context.Context) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
	fmt.Println("cleanup")
	return nil
}); err != nil {
	log.Fatal(err)
}

// Main process of some kind

defer func() {
	cancel()
	timeout := 5 * time.Millisecond
	if err := WaitWithTimeout(ctx, timeout); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
}()

fmt.Println("finish")

// Output:
// finish
```

