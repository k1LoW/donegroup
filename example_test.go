package donegroup_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/k1LoW/donegroup"
)

func Example() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	// Cleanup process of some kind
	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("cleanup with sleep")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Cleanup process of some kind
	if err := donegroup.Clenup(ctx, func(_ context.Context) error {
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

func ExampleWaitWithTimeout() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	// Cleanup process of some kind
	if err := donegroup.Clenup(ctx, func(ctx context.Context) error {
		fmt.Println("cleanup start")
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(2 * time.Millisecond)
			}
		}
		fmt.Println("cleanup end")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Main process of some kind

	defer func() {
		cancel()
		timeout := 5 * time.Millisecond
		if err := donegroup.WaitWithTimeout(ctx, timeout); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			log.Fatal(err)
		}
	}()

	fmt.Println("finish")

	// Output:
	// finish
	// cleanup start
}
