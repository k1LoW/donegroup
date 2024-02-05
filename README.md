# donegroup

`donegroup` is a package that provides a graceful cleanup transaction to context.Context when the context is canceled.

## Usage

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

	if err := donegroup.Clenup(ctx, func() error {
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
