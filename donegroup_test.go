package donegroup

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoneGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	defer func() {
		cancel()

		if err := Wait(ctx); err != nil {
			t.Error(err)
		}

		if !cleanup.Load() {
			t.Error("cleanup function not called")
		}
	}()

	cleanup.Store(false)
}

func TestCleanup(t *testing.T) {
	t.Parallel()
	t.Run("Cleanup with WithCancel", func(t *testing.T) {
		ctx, cancel := WithCancel(context.Background())
		defer cancel()
		err := Cleanup(ctx, func() error {
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Cleanup without WithCancel", func(t *testing.T) {
		ctx := context.Background()
		err := Cleanup(ctx, func() error {
			return nil
		})
		if !errors.Is(err, ErrNotContainDoneGroup) {
			t.Errorf("expected ErrNotContainDoneGroup, got %v", err)
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
		if !errors.Is(err, ErrNotContainDoneGroup) {
			t.Errorf("expected ErrNotContainDoneGroup, got %v", err)
		}
	})

	t.Run("Collect errors", func(t *testing.T) {
		var (
			errTest  = errors.New("test error")
			errTest2 = errors.New("test error 2")
		)

		ctx, cancel := WithCancel(context.Background())
		if err := Cleanup(ctx, func() error {
			return errTest
		}); err != nil {
			t.Error(err)
		}
		if err := Cleanup(ctx, func() error {
			return errTest2
		}); err != nil {
			t.Error(err)
		}
		cancel()
		err := Wait(ctx)
		if !errors.Is(err, errTest) {
			t.Errorf("expected %v, got %v", errTest, err)
		}
		if !errors.Is(err, errTest2) {
			t.Errorf("expected %v, got %v", errTest2, err)
		}
	})
}

func TestNoWait(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	defer func() {
		cancel()
		if cleanup.Load() {
			t.Error("cleanup function called")
		}

		time.Sleep(20 * time.Millisecond)
		if !cleanup.Load() {
			t.Error("cleanup function not called")
		}
	}()

	cleanup.Store(false)
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

	cleanup := atomic.Int64{}

	for i := 0; i < 10; i++ {
		if err := Cleanup(ctx, func() error {
			time.Sleep(10 * time.Millisecond)
			cleanup.Add(1)
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

		if cleanup.Load() != 10 {
			t.Error("cleanup function not called")
		}
	}()
}

func TestNestedWithCancel(t *testing.T) {
	t.Parallel()
	firstCtx, firstCancel := WithCancel(context.Background())
	secondCtx, secondCancel := WithCancel(firstCtx)
	thirdCtx, thirdCancel := context.WithCancel(secondCtx) // context.WithCancel

	firstCleanup := atomic.Int64{}
	secondCleanup := atomic.Int64{}
	thirdCleanup := atomic.Int64{}

	for i := 0; i < 10; i++ {
		if err := Cleanup(firstCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			firstCleanup.Add(1)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Cleanup(secondCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			secondCleanup.Add(1)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 3; i++ {
		if err := Cleanup(thirdCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			thirdCleanup.Add(1)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	{
		dg, ok := firstCtx.Value(doneGroupKey).(*doneGroup)
		if !ok {
			t.Fatal("firstCtx.Value(doneGroupKey) is not *doneGroup")
		}
		got := len(dg.cleanupGroups)
		if want := 2; got != want {
			t.Errorf("firstCtx has %d cleanup groups, want %d", got, want)
		}
	}

	{
		dg, ok := secondCtx.Value(doneGroupKey).(*doneGroup)
		if !ok {
			t.Fatal("firstCtx.Value(doneGroupKey) is not *doneGroup")
		}
		got := len(dg.cleanupGroups)
		if want := 1; got != want {
			t.Errorf("firstCtx has %d cleanup groups, want %d", got, want)
		}
	}

	defer func() {
		thirdCancel()
		<-thirdCtx.Done()

		if firstCleanup.Load() != 0 {
			t.Error("cleanup function for first called")
		}
		if secondCleanup.Load() != 0 {
			t.Error("cleanup function for second called")
		}
		if thirdCleanup.Load() != 0 {
			t.Error("cleanup function for third called")
		}

		secondCancel()
		<-secondCtx.Done()

		if err := Wait(secondCtx); err != nil {
			t.Error(err)
		}

		if thirdCleanup.Load() != 3 {
			t.Error("cleanup function for third not called")
		}
		if secondCleanup.Load() != 5 {
			t.Error("cleanup function for second not called")
		}
		if firstCleanup.Load() != 0 {
			t.Error("cleanup function for first called")
		}

		firstCancel()
		<-firstCtx.Done()

		if err := Wait(firstCtx); err != nil {
			t.Error(err)
		}

		if thirdCleanup.Load() != 3 {
			t.Error("cleanup function for third not called")
		}
		if secondCleanup.Load() != 5 {
			t.Error("cleanup function for second not called")
		}
		if firstCleanup.Load() != 10 {
			t.Error("cleanup function for first not called")
		}
	}()
}

func TestRootWaitAll(t *testing.T) {
	t.Parallel()
	rootCtx, rootCancel := WithCancel(context.Background())
	leafCtx, _ := WithCancel(rootCtx)

	rootCleanup := atomic.Int64{}
	leafCleanup := atomic.Int64{}

	for i := 0; i < 10; i++ {
		if err := Cleanup(rootCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			rootCleanup.Add(1)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Cleanup(leafCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			leafCleanup.Add(1)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	defer func() {
		if rootCleanup.Load() != 0 {
			t.Error("cleanup function for root called")
		}

		rootCancel()

		if err := Wait(rootCtx); err != nil {
			t.Error(err)
		}

		if leafCleanup.Load() != 5 {
			t.Error("cleanup function for leaf not called")
		}

		if rootCleanup.Load() != 10 {
			t.Error("cleanup function for root not called")
		}
	}()
}

func TestWaitWithTimeout(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	if err := Cleanup(ctx, func() error {
		for i := 0; i < 10; i++ {
			time.Sleep(2 * time.Millisecond)
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

	if err := Cleanup(ctx, func() error {
		for i := 0; i < 10; i++ {
			time.Sleep(2 * time.Millisecond)
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

			var finished int32

			go func() {
				completed, err := Awaiter(ctx)
				if err != nil {
					t.Error(err)
				}
				<-ctx.Done()
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&finished, 1)
				completed()
			}()

			defer func() {
				cancel()
				time.Sleep(10 * time.Millisecond)
				err := WaitWithTimeout(ctx, tt.timeout)
				if tt.finished != (atomic.LoadInt32(&finished) > 0) {
					t.Errorf("expected finished: %v, got: %v", tt.finished, finished)
				}
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

			var finished int32

			go func() {
				defer Awaitable(ctx)()
				<-ctx.Done()
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&finished, 1)
			}()

			defer func() {
				cancel()
				time.Sleep(10 * time.Millisecond)
				err := WaitWithTimeout(ctx, tt.timeout)
				if tt.finished != (atomic.LoadInt32(&finished) > 0) {
					t.Errorf("expected finished: %v, got: %v", tt.finished, finished)
				}
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
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Error("expected context.Canceled")
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

func TestGo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		timeout  time.Duration
		finished bool
	}{
		{
			name:     "finished",
			timeout:  200 * time.Millisecond,
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

			var finished int32

			Go(ctx, func() error {
				<-ctx.Done()
				time.Sleep(100 * time.Millisecond)
				atomic.AddInt32(&finished, 1)
				return nil
			})

			defer func() {
				cancel()
				time.Sleep(10 * time.Millisecond)
				err := WaitWithTimeout(ctx, tt.timeout)
				if tt.finished != (atomic.LoadInt32(&finished) > 0) {
					t.Errorf("expected finished: %v, got: %v", tt.finished, finished)
				}
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

func TestGoWithError(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())

	var errTest = errors.New("test error")

	Go(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		return errTest
	})

	defer func() {
		cancel()

		err := Wait(ctx)
		if !errors.Is(err, errTest) {
			t.Errorf("got %v, want %v", err, errTest)
		}
	}()
}

func TestWithCancelCause(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancelCause(context.Background())

	cleanup := false

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup = true
		return nil
	}); err != nil {
		t.Error(err)
	}

	var errTest = errors.New("test error")

	defer func() {
		cancel(errTest)

		if err := Wait(ctx); err != nil {
			t.Error(err)
		}

		if !cleanup {
			t.Error("cleanup function not called")
		}

		if !errors.Is(context.Cause(ctx), errTest) {
			t.Errorf("got %v, want %v", context.Cause(ctx), errTest)
		}
	}()

	cleanup = false
}

func TestWithDeadline(t *testing.T) {
	t.Parallel()
	ctx, _ := WithDeadline(context.Background(), time.Now().Add(5*time.Millisecond))

	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	cleanup.Store(false)

	if err := Wait(ctx); err != nil {
		t.Error(err)
	}

	if !cleanup.Load() {
		t.Error("cleanup function not called")
	}

	if !errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
		t.Errorf("got %v, want %v", context.Cause(ctx), context.DeadlineExceeded)
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()
	ctx, _ := WithTimeout(context.Background(), 5*time.Millisecond)

	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	cleanup.Store(false)

	if err := Wait(ctx); err != nil {
		t.Error(err)
	}

	if !cleanup.Load() {
		t.Error("cleanup function not called")
	}

	if !errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
		t.Errorf("got %v, want %v", context.Cause(ctx), context.DeadlineExceeded)
	}
}

func TestWithTimeoutCause(t *testing.T) {
	t.Parallel()
	var errTest = errors.New("test error")
	ctx, _ := WithTimeoutCause(context.Background(), 5*time.Millisecond, errTest)

	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		time.Sleep(10 * time.Millisecond)
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	cleanup.Store(false)

	if err := Wait(ctx); err != nil {
		t.Error(err)
	}

	if !cleanup.Load() {
		t.Error("cleanup function not called")
	}

	if !errors.Is(context.Cause(ctx), errTest) {
		t.Errorf("got %v, want %v", context.Cause(ctx), errTest)
	}
}

func TestCancelWithCause(t *testing.T) {
	t.Parallel()
	var errTest = errors.New("test error")
	var errTest2 = errors.New("test error2")

	t.Run("Cancel with cause", func(t *testing.T) {
		ctx, _ := WithCancel(context.Background())

		if err := CancelWithCause(ctx, errTest); err != nil {
			t.Error(err)
		}

		if !errors.Is(context.Cause(ctx), errTest) {
			t.Errorf("got %v, want %v", context.Cause(ctx), errTest)
		}
	})

	t.Run("Cancel with cause2", func(t *testing.T) {
		ctx, _ := WithCancel(context.Background())

		if err := CancelWithCause(ctx, errTest); err != nil {
			t.Error(err)
		}

		if err := CancelWithCause(ctx, errTest2); err != nil {
			t.Error(err)
		}

		if !errors.Is(context.Cause(ctx), errTest) {
			t.Errorf("got %v, want %v", context.Cause(ctx), errTest)
		}
		if errors.Is(context.Cause(ctx), errTest2) {
			t.Error("got errTest2, want errTest")
		}
	})
}

func TestWithoutCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := WithCancel(context.Background())
	cleanup := atomic.Bool{}

	if err := Cleanup(ctx, func() error {
		cleanup.Store(true)
		return nil
	}); err != nil {
		t.Error(err)
	}

	ctx = WithoutCancel(ctx)

	if err := Wait(ctx); err == nil || !errors.Is(err, ErrNotContainDoneGroup) {
		t.Errorf("got %v, want %v", err, ErrNotContainDoneGroup)
	}

	cancel()

	time.Sleep(5 * time.Millisecond)

	if !cleanup.Load() {
		t.Error("cleanup function not called")
	}
}
