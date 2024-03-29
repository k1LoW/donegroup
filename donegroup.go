package donegroup

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var doneGroupKey = struct{}{}
var ErrNotContainDoneGroup = errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")

// doneGroup is cleanup function groups per Context.
type doneGroup struct {
	cancel context.CancelFunc
	// ctxw is the context used to call the cleanup functions
	ctxw          context.Context
	cleanupGroups []*errgroup.Group
	errors        error
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
		return ErrNotContainDoneGroup
	}

	first := dg.cleanupGroups[0]
	first.Go(func() error {
		<-ctx.Done()
		<-dg._ctx.Done()
		dg.mu.Lock()
		ctxw := dg.ctxw
		dg.mu.Unlock()
		if err := f(ctxw); err != nil {
			dg.mu.Lock()
			dg.errors = errors.Join(dg.errors, err)
			dg.mu.Unlock()
		}
		return nil
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

// Cancel cancels the context. Then calls the function registered by Cleanup.
func Cancel(ctx context.Context) error {
	return CancelWithKey(ctx, doneGroupKey)
}

// CancelWithTimeout cancels the context. Then calls the function registered by Cleanup with timeout.
func CancelWithTimeout(ctx context.Context, timeout time.Duration) error {
	return CancelWithTimeoutAndKey(ctx, timeout, doneGroupKey)
}

// CancelWithContext cancels the context. Then calls the function registered by Cleanup with context (ctxw).
func CancelWithContext(ctx, ctxw context.Context) error {
	return CancelWithContextAndKey(ctx, ctxw, doneGroupKey)
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
		return ErrNotContainDoneGroup
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
	if err := eg.Wait(); err != nil {
		return errors.Join(dg.errors, err)
	}
	return dg.errors
}

// CancelWithKey cancels the context. Then calls the function registered by Cleanup.
func CancelWithKey(ctx context.Context, key any) error {
	return CancelWithContextAndKey(ctx, context.Background(), key)
}

// CancelWithTimeoutAndKey cancels the context. Then calls the function registered by Cleanup with timeout.
func CancelWithTimeoutAndKey(ctx context.Context, timeout time.Duration, key any) error {
	ctxw, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return CancelWithContextAndKey(ctx, ctxw, key)
}

// CancelWithContextAndKey cancels the context. Then calls the function registered by Cleanup with context (ctxw).
func CancelWithContextAndKey(ctx, ctxw context.Context, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return ErrNotContainDoneGroup
	}
	dg.cancel()
	return WaitWithContextAndKey(ctx, ctxw, key)
}

// Awaiter returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func Awaiter(ctx context.Context) (completed func(), err error) {
	return AwaiterWithKey(ctx, doneGroupKey)
}

// AwaiterWithKey returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func AwaiterWithKey(ctx context.Context, key any) (completed func(), err error) {
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
		return nil, err
	}
	return completed, nil
}

// Awaitable returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func Awaitable(ctx context.Context) (completed func()) {
	return AwaitableWithKey(ctx, doneGroupKey)
}

// AwaitableWithKey returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func AwaitableWithKey(ctx context.Context, key any) (completed func()) {
	completed, err := AwaiterWithKey(ctx, key)
	if err != nil {
		panic(err)
	}
	return completed
}
