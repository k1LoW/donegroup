package donegroup

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestDoneGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	cleanup := false

	if err := Cleanup(ctx, func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		cleanup = true
		return nil
	}); err != nil {
		t.Error(err)
	}

	defer func() {
		cancel()

		if err := Wait(ctx); err != nil {
			t.Error(err)
		}

		if !cleanup {
			t.Error("cleanup function not called")
		}
	}()

	cleanup = false
}

func TestCleanup(t *testing.T) {
	t.Parallel()
	t.Run("Cleanup with WithCancel", func(t *testing.T) {
		ctx, cancel := WithCancel(context.Background())
		defer cancel()
		err := Cleanup(ctx, func(_ context.Context) error {
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Cleanup without WithCancel", func(t *testing.T) {
		ctx := context.Background()
		err := Cleanup(ctx, func(_ context.Context) error {
			return nil
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWait(t *testing.T) {
	t.Parallel()
	t.Run("Wait with WithCancel", func(t *testing.T) {
		ctx, cancel := WithCancel(context.Background())
		cancel()
		err := Wait(ctx)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Wait without WithCancel", func(t *testing.T) {
		ctx := context.Background()
		err := Wait(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestMultiCleanup(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	mu := sync.Mutex{}
	cleanup := 0

	for i := 0; i < 10; i++ {
		if err := Cleanup(ctx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			cleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	defer func() {
		cancel()

		if err := Wait(ctx); err != nil {
			t.Error(err)
		}

		if cleanup != 10 {
			t.Error("cleanup function not called")
		}
	}()
}

func TestNested(t *testing.T) {
	t.Parallel()
	firstCtx, firstCancel := WithCancel(context.Background())
	secondCtx, secondCancel := WithCancel(firstCtx)

	mu := sync.Mutex{}
	firstCleanup := 0
	secondCleanup := 0

	for i := 0; i < 10; i++ {
		if err := Cleanup(firstCtx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			firstCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Cleanup(secondCtx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			secondCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	defer func() {
		secondCancel()

		if err := Wait(secondCtx); err != nil {
			t.Error(err)
		}

		if secondCleanup != 5 {
			t.Error("cleanup function for second not called")
		}
		if firstCleanup != 0 {
			t.Error("cleanup function for first called")
		}

		firstCancel()

		if err := Wait(firstCtx); err != nil {
			t.Error(err)
		}

		if firstCleanup != 10 {
			t.Error("cleanup function for first not called")
		}
	}()
}

func TestRootWaitAll(t *testing.T) {
	t.Parallel()
	rootCtx, rootCancel := WithCancel(context.Background())
	leafCtx, _ := WithCancel(rootCtx)

	mu := sync.Mutex{}
	rootCleanup := 0
	leafCleanup := 0

	for i := 0; i < 10; i++ {
		if err := Cleanup(rootCtx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			rootCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Cleanup(leafCtx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			leafCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	defer func() {
		if rootCleanup != 0 {
			t.Error("cleanup function for root called")
		}

		rootCancel()

		if err := Wait(rootCtx); err != nil {
			t.Error(err)
		}

		if leafCleanup != 5 {
			t.Error("cleanup function for leaf not called")
		}

		if rootCleanup != 10 {
			t.Error("cleanup function for root not called")
		}
	}()
}

func TestWaitWithTimeout(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	if err := Cleanup(ctx, func(ctx context.Context) error {
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(2 * time.Millisecond)
			}
		}
		return nil
	}); err != nil {
		t.Error(err)
	}

	timeout := 5 * time.Millisecond

	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
		if err := WaitWithTimeout(ctx, timeout); !errors.Is(err, context.DeadlineExceeded) {
			t.Error("expected timeout error")
		}
	}()
}

func TestWaitWithContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	if err := Cleanup(ctx, func(ctx context.Context) error {
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(2 * time.Millisecond)
			}
		}
		return nil
	}); err != nil {
		t.Error(err)
	}

	timeout := 5 * time.Millisecond

	defer func() {
		cancel()
		ctxx, cancelx := context.WithTimeout(context.Background(), timeout)
		defer cancelx()
		time.Sleep(10 * time.Millisecond)
		if err := WaitWithContext(ctx, ctxx); !errors.Is(err, context.DeadlineExceeded) {
			t.Error("expected timeout error")
		}
	}()
}
