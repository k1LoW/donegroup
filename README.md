# donegroup

`donegroup` is a package that provides a graceful cleanup transaction to context.Context when the context is canceled ( **done** ).

> errgroup.Group after <-ctx.Done() = donegroup

## Usage

Use donegroup.WithCancel instead of [context.WithCancel](https://pkg.go.dev/context#WithCancel).

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
		fmt.Println("cleanup")
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
	// cleanup
}
```

dongroup.Cleanup is similar in usage to [testing.(*T) Cleanup](https://pkg.go.dev/testing#T.Cleanup), but the order of execution is not guaranteed.
