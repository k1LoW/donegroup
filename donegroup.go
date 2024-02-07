package donegroup

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var doneGroupKey = struct{}{}

// doneGroup is cleanup function groups per Context.
type doneGroup struct {
	cancel context.CancelFunc
	// ctxw is the context used to call the cleanup functions
	ctxw          context.Context
	cleanupGroups []*errgroup.Group
	mu            sync.Mutex
	// _ctx, _cancel is a context/cancelFunc used to set dg.ctxw
	_ctx    context.Context
	_cancel context.CancelFunc
}

// WithCancel returns a copy of parent with a new Done channel and a doneGroup.
func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithCancelWithKey(ctx, doneGroupKey)
}

// WithCancelWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithCancelWithKey(ctx context.Context, key any) (context.Context, context.CancelFunc) {
	dgCtx, dgCancel := context.WithCancel(ctx)
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		dg = &doneGroup{
			cancel:  dgCancel,
			_ctx:    ctx,
			_cancel: cancel,
		}
	}
	eg := new(errgroup.Group)
	dg.cleanupGroups = append(dg.cleanupGroups, eg)
	secondDg := &doneGroup{
		cancel:        dgCancel,
		_ctx:          dg._ctx,
		_cancel:       dg._cancel,
		cleanupGroups: []*errgroup.Group{eg},
	}
	return context.WithValue(dgCtx, key, secondDg), dgCancel
}

// Cleanup registers a function to be called when the context is canceled.
func Cleanup(ctx context.Context, f func(ctx context.Context) error) error {
	return CleanupWithKey(ctx, doneGroupKey, f)
}

// CleanupWithKey Cleanup registers a function to be called when the context is canceled.
func CleanupWithKey(ctx context.Context, key any, f func(ctx context.Context) error) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")
	}

	first := dg.cleanupGroups[0]
	first.Go(func() error {
		<-ctx.Done()
		<-dg._ctx.Done()
		dg.mu.Lock()
		ctxw := dg.ctxw
		dg.mu.Unlock()
		return f(ctxw)
	})
	return nil
}

// Wait blocks until the context is canceled. Then calls the function registered by Cleanup.
func Wait(ctx context.Context) error {
	return WaitWithKey(ctx, doneGroupKey)
}

// WaitWithTimeout blocks until the context (ctx) is canceled. Then calls the function registered by Cleanup with timeout.
func WaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	return WaitWithTimeoutAndKey(ctx, timeout, doneGroupKey)
}

// WaitWithContext blocks until the context (ctx) is canceled. Then calls the function registered by Cleanup with context (ctxw).
func WaitWithContext(ctx, ctxw context.Context) error {
	return WaitWithContextAndKey(ctx, ctxw, doneGroupKey)
}

// CancelAndWait cancels the context. Then calls the function registered by Cleanup.
func CancelAndWait(ctx context.Context) error {
	return CancelAndWaitWithKey(ctx, doneGroupKey)
}

// CancelAndWaitWithTimeout cancels the context. Then calls the function registered by Cleanup with timeout.
func CancelAndWaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	return CancelAndWaitWithTimeoutAndKey(ctx, timeout, doneGroupKey)
}

// CancelAndWaitWithContext cancels the context. Then calls the function registered by Cleanup with context (ctxw).
func CancelAndWaitWithContext(ctx, ctxw context.Context) error {
	return CancelAndWaitWithContextAndKey(ctx, ctxw, doneGroupKey)
}

// WaitWithKey blocks until the context is canceled. Then calls the function registered by Cleanup.
func WaitWithKey(ctx context.Context, key any) error {
	return WaitWithContextAndKey(ctx, context.Background(), key)
}

// WaitWithTimeoutAndKey blocks until the context is canceled. Then calls the function registered by Cleanup with timeout.
func WaitWithTimeoutAndKey(ctx context.Context, timeout time.Duration, key any) error {
	ctxw, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WaitWithContextAndKey(ctx, ctxw, key)
}

// WaitWithContextAndKey blocks until the context is canceled. Then calls the function registered by Cleanup with context (ctxx).
func WaitWithContextAndKey(ctx, ctxw context.Context, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")
	}
	dg.mu.Lock()
	dg.ctxw = ctxw
	dg.mu.Unlock()
	<-ctx.Done()
	dg._cancel()
	eg, _ := errgroup.WithContext(ctxw)
	for _, g := range dg.cleanupGroups {
		eg.Go(g.Wait)
	}

	return eg.Wait()
}

// CancelAndWaitWithKey cancels the context. Then calls the function registered by Cleanup.
func CancelAndWaitWithKey(ctx context.Context, key any) error {
	return CancelAndWaitWithContextAndKey(ctx, context.Background(), key)
}

// CancelAndWaitWithTimeoutAndKey cancels the context. Then calls the function registered by Cleanup with timeout.
func CancelAndWaitWithTimeoutAndKey(ctx context.Context, timeout time.Duration, key any) error {
	ctxw, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return CancelAndWaitWithContextAndKey(ctx, ctxw, key)
}

// CancelAndWaitWithContextAndKey cancels the context. Then calls the function registered by Cleanup with context (ctxw).
func CancelAndWaitWithContextAndKey(ctx, ctxw context.Context, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")
	}
	dg.cancel()
	return WaitWithContextAndKey(ctx, ctxw, key)
}

// Awaiter returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func Awaiter(ctx context.Context) (completed func()) {
	return AwaiterWithKey(ctx, doneGroupKey)
}

// AwaiterWithKey returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func AwaiterWithKey(ctx context.Context, key any) (completed func()) {
	ctxx, completed := context.WithCancel(context.Background())
	if err := CleanupWithKey(ctx, key, func(ctxw context.Context) error {
		for {
			select {
			case <-ctxw.Done():
				return ctxw.Err()
			case <-ctxx.Done():
				return nil
			}
		}
	}); err != nil {
		panic(err)
	}
	return completed
}
