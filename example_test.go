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

	// Cleanup process of some kind
	if err := donegroup.Cleanup(ctx, func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("cleanup with sleep")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Cleanup process of some kind
	if err := donegroup.Cleanup(ctx, func(_ context.Context) error {
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

	// Main process of some kind
	fmt.Println("main start")

	fmt.Println("main finish")

	// Output:
	// main start
	// main finish
	// cleanup
	// cleanup with sleep
}

func Example_goroutine() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	go func() {
		if err := donegroup.Cleanup(ctx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("cleanup")
			return nil
		}); err != nil {
			log.Fatal(err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
				fmt.Println("do something")
			}
		}
	}()

	// Main process of some kind
	fmt.Println("main")
	time.Sleep(35 * time.Millisecond)

	cancel()
	if err := donegroup.Wait(ctx); err != nil {
		log.Fatal(err)
	}

	// Output:
	// main
	// do something
	// do something
	// do something
	// cleanup
}

func ExampleWaitWithTimeout() {
	ctx, cancel := donegroup.WithCancel(context.Background())

	// Cleanup process of some kind
	if err := donegroup.Cleanup(ctx, func(ctx context.Context) error {
		fmt.Println("cleanup start")
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(2 * time.Millisecond)
			}
		}
		fmt.Println("cleanup finish")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	defer func() {
		cancel()
		timeout := 5 * time.Millisecond
		if err := donegroup.WaitWithTimeout(ctx, timeout); err != nil {
			fmt.Println(err)
		}
	}()

	// Main process of some kind
	fmt.Println("main start")

	fmt.Println("main finish")

	// Output:
	// main start
	// main finish
	// cleanup start
	// context deadline exceeded
}
