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

func TestNoCleanup(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	defer func() {
		cancel()

		if err := Wait(ctx); err != nil {
			t.Error(err)
		}
	}()
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

func TestNestedWithCancel(t *testing.T) {
	t.Parallel()
	firstCtx, firstCancel := WithCancel(context.Background())
	secondCtx, secondCancel := WithCancel(firstCtx)
	thirdCtx, thirdCancel := context.WithCancel(secondCtx) // context.WithCancel

	mu := sync.Mutex{}
	firstCleanup := 0
	secondCleanup := 0
	thirdCleanup := 0

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

	for i := 0; i < 3; i++ {
		if err := Cleanup(thirdCtx, func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			defer mu.Unlock()
			thirdCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	defer func() {
		thirdCancel()
		<-thirdCtx.Done()

		if firstCleanup != 0 {
			t.Error("cleanup function for first called")
		}
		if secondCleanup != 0 {
			t.Error("cleanup function for second called")
		}
		if thirdCleanup != 0 {
			t.Error("cleanup function for third called")
		}

		secondCancel()
		<-secondCtx.Done()

		if err := Wait(secondCtx); err != nil {
			t.Error(err)
		}

		if thirdCleanup != 3 {
			t.Error("cleanup function for third not called")
		}
		if secondCleanup != 5 {
			t.Error("cleanup function for second not called")
		}
		if firstCleanup != 0 {
			t.Error("cleanup function for first called")
		}

		firstCancel()
		<-firstCtx.Done()

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

func TestAwaiter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		timeout  time.Duration
		finished bool
	}{
		{
			name:     "finished",
			timeout:  100 * time.Millisecond,
			finished: true,
		},
		{
			name:     "not finished",
			timeout:  5 * time.Millisecond,
			finished: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := WithCancel(context.Background())

			go func() {
				completed, err := Awaiter(ctx)
				if err != nil {
					t.Error(err)
				}
				<-ctx.Done()
				time.Sleep(20 * time.Millisecond)
				completed()
			}()

			defer func() {
				cancel()
				time.Sleep(10 * time.Millisecond)
				err := WaitWithTimeout(ctx, tt.timeout)
				if tt.finished {
					if err != nil {
						t.Error(err)
					}
					return
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("expected timeout error: %v", err)
				}
			}()
		})
	}
}

func TestAwaitable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		timeout  time.Duration
		finished bool
	}{
		{
			name:     "finished",
			timeout:  100 * time.Millisecond,
			finished: true,
		},
		{
			name:     "not finished",
			timeout:  5 * time.Millisecond,
			finished: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := WithCancel(context.Background())

			go func() {
				defer Awaitable(ctx)()
				<-ctx.Done()
				time.Sleep(20 * time.Millisecond)
			}()

			defer func() {
				cancel()
				time.Sleep(10 * time.Millisecond)
				err := WaitWithTimeout(ctx, tt.timeout)
				if tt.finished {
					if err != nil {
						t.Error(err)
					}
					return
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("expected timeout error: %v", err)
				}
			}()
		})
	}
}

func TestCancel(t *testing.T) {
	t.Parallel()
	t.Run("Cancel with WithCancel", func(t *testing.T) {
		ctx, _ := WithCancel(context.Background())
		err := Cancel(ctx)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Cancel without WithCancel", func(t *testing.T) {
		ctx := context.Background()
		err := Cancel(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCancelWithTimeout(t *testing.T) {
	t.Parallel()
	ctx, _ := WithCancel(context.Background())

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
		time.Sleep(10 * time.Millisecond)
		if err := CancelWithTimeout(ctx, timeout); !errors.Is(err, context.DeadlineExceeded) {
			t.Error("expected timeout error")
		}
	}()
}

func TestCancelWithContext(t *testing.T) {
	t.Parallel()
	ctx, _ := WithCancel(context.Background())

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
		ctxx, cancelx := context.WithTimeout(context.Background(), timeout)
		defer cancelx()
		time.Sleep(10 * time.Millisecond)
		if err := CancelWithContext(ctx, ctxx); !errors.Is(err, context.DeadlineExceeded) {
			t.Error("expected timeout error")
		}
	}()
}
