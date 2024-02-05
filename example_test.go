package donegroup_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/k1LoW/donegroup"
)

func Example() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		// Cleanup process of some kind
		time.Sleep(10 * time.Millisecond)
		fmt.Println("cleanup with sleep")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		// Cleanup process of some kind
		fmt.Println("cleanup")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

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
	// cleanup with sleep
}
