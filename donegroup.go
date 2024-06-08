package donegroup

import (
	"context"
	"errors"
	"sync"
	"time"
)

var doneGroupKey = struct{}{}
var ErrNotContainDoneGroup = errors.New("donegroup: context does not contain a doneGroup. Use donegroup.With* to create a context with a doneGroup")

// doneGroup is cleanup function groups per Context.
type doneGroup struct {
	cancel        context.CancelCauseFunc
	cleanupGroups []*sync.WaitGroup
	errors        error
	mu            sync.Mutex
}

// WithCancel returns a copy of parent with a new Done channel and a doneGroup.
func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithCancelWithKey(ctx, doneGroupKey)
}

// WithDeadline returns a copy of parent with a new Done channel and a doneGroup.
// If the deadline is exceeded, the cause is set to context.DeadlineExceeded.
func WithDeadline(ctx context.Context, d time.Time) (context.Context, context.CancelFunc) {
	return WithDeadlineCause(ctx, d, nil)
}

// WithTimeout returns a copy of parent with a new Done channel and a doneGroup.
// If the timeout is exceeded, the cause is set to context.DeadlineExceeded.
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return WithTimeoutCause(ctx, timeout, nil)
}

// WithCancelCause returns a copy of parent with a new Done channel and a doneGroup.
func WithCancelCause(ctx context.Context) (context.Context, context.CancelCauseFunc) {
	return WithCancelCauseWithKey(ctx, doneGroupKey)
}

// WithDeadlineCause returns a copy of parent with a new Done channel and a doneGroup.
func WithDeadlineCause(ctx context.Context, d time.Time, cause error) (context.Context, context.CancelFunc) {
	return WithDeadlineCauseWithKey(ctx, d, cause, doneGroupKey)
}

// WithTimeoutCause returns a copy of parent with a new Done channel and a doneGroup.
func WithTimeoutCause(ctx context.Context, timeout time.Duration, cause error) (context.Context, context.CancelFunc) {
	return WithTimeoutCauseWithKey(ctx, timeout, cause, doneGroupKey)
}

// WithoutCancel returns a copy of parent that is not canceled when parent is canceled and does not have a doneGroup.
func WithoutCancel(ctx context.Context) context.Context {
	return WithoutCancelWithKey(ctx, doneGroupKey)
}

// WithoutCancelWithKey returns a copy of parent that is not canceled when parent is canceled and does not have a doneGroup.
func WithoutCancelWithKey(ctx context.Context, key any) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), key, nil)
}

// WithCancelWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithCancelWithKey(ctx context.Context, key any) (context.Context, context.CancelFunc) {
	ctx, cancelCause := WithCancelCauseWithKey(ctx, key)
	return ctx, func() { cancelCause(nil) }
}

// WithDeadlineWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithDeadlineWithKey(ctx context.Context, d time.Time, key any) (context.Context, context.CancelFunc) {
	return WithDeadlineCauseWithKey(ctx, d, nil, key)
}

// WithTimeoutWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithTimeoutWithKey(ctx context.Context, timeout time.Duration, key any) (context.Context, context.CancelFunc) {
	return WithTimeoutCauseWithKey(ctx, timeout, nil, key)
}

// WithCancelCauseWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithCancelCauseWithKey(ctx context.Context, key any) (context.Context, context.CancelCauseFunc) {
	ctx, cancelCause := context.WithCancelCause(ctx)
	return withDoneGroup(ctx, cancelCause, key), cancelCause
}

// WithDeadlineCauseWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithDeadlineCauseWithKey(ctx context.Context, d time.Time, cause error, key any) (context.Context, context.CancelFunc) {
	ctx, cancelCause := context.WithCancelCause(ctx)
	ctx, cancel := context.WithDeadlineCause(ctx, d, cause)
	ctx = withDoneGroup(ctx, cancelCause, key)
	return ctx, cancel
}

// WithTimeoutCauseWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithTimeoutCauseWithKey(ctx context.Context, timeout time.Duration, cause error, key any) (context.Context, context.CancelFunc) {
	return WithDeadlineCauseWithKey(ctx, time.Now().Add(timeout), cause, key)
}

// Cleanup registers a function to be called when the context is canceled.
func Cleanup(ctx context.Context, f func() error) error {
	return CleanupWithKey(ctx, doneGroupKey, f)
}

// CleanupWithKey Cleanup registers a function to be called when the context is canceled.
func CleanupWithKey(ctx context.Context, key any, f func() error) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return ErrNotContainDoneGroup
	}

	rootWg := dg.cleanupGroups[0]
	dg.mu.Lock()
	rootWg.Add(1)
	dg.mu.Unlock()

	_ = context.AfterFunc(ctx, func() {
		if err := f(); err != nil {
			dg.mu.Lock()
			dg.errors = errors.Join(dg.errors, err)
			dg.mu.Unlock()
		}
		rootWg.Done()
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

// CancelWithCause cancels the context with cause. Then calls the function registered by Cleanup.
func CancelWithCause(ctx context.Context, cause error) error {
	return CancelWithCauseAndKey(ctx, cause, doneGroupKey)
}

// WaitWithKey blocks until the context is canceled. Then calls the function registered by Cleanup.
func WaitWithKey(ctx context.Context, key any) error {
	return WaitWithContextAndKey(ctx, context.WithoutCancel(ctx), key)
}

// WaitWithTimeoutAndKey blocks until the context is canceled. Then calls the function registered by Cleanup with timeout.
func WaitWithTimeoutAndKey(ctx context.Context, timeout time.Duration, key any) error {
	ctxw, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	return WaitWithContextAndKey(ctx, ctxw, key)
}

// WaitWithContextAndKey blocks until the context is canceled. Then calls the function registered by Cleanup with context (ctxx).
func WaitWithContextAndKey(ctx, ctxw context.Context, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return ErrNotContainDoneGroup
	}
	<-ctx.Done()
	wg := &sync.WaitGroup{}
	for _, g := range dg.cleanupGroups {
		wg.Add(1)
		dg.mu.Lock()
		go func() {
			g.Wait()
			wg.Done()
		}()
		dg.mu.Unlock()
	}
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
	case <-ctxw.Done():
		dg.mu.Lock()
		defer dg.mu.Unlock()
		dg.errors = errors.Join(dg.errors, ctxw.Err())
	}
	return dg.errors
}

// CancelWithKey cancels the context.
func CancelWithKey(ctx context.Context, key any) error {
	return CancelWithCauseAndKey(ctx, nil, key)
}

// CancelWithCauseAndKey cancels the context with cause.
func CancelWithCauseAndKey(ctx context.Context, cause error, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return ErrNotContainDoneGroup
	}
	dg.cancel(cause)
	return nil
}

// Awaiter returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func Awaiter(ctx context.Context) (completed func(), err error) {
	return AwaiterWithKey(ctx, doneGroupKey)
}

// AwaiterWithKey returns a function that guarantees execution of the process until it is called.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func AwaiterWithKey(ctx context.Context, key any) (completed func(), err error) {
	ctxx, completed := context.WithCancel(context.WithoutCancel(ctx)) //nolint:govet
	if err := CleanupWithKey(ctx, key, func() error {
		<-ctxx.Done()
		return nil
	}); err != nil {
		return nil, err //nolint:govet
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

// Go calls the function now asynchronously.
// If an error occurs, it is stored in the doneGroup.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func Go(ctx context.Context, f func() error) {
	GoWithKey(ctx, doneGroupKey, f)
}

// GoWithKey calls the function now asynchronously.
// If an error occurs, it is stored in the doneGroup.
// Note that if the timeout of WaitWithTimeout has passed (or the context of WaitWithContext has canceled), it will not wait.
func GoWithKey(ctx context.Context, key any, f func() error) {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		panic(ErrNotContainDoneGroup)
	}
	completed, err := AwaiterWithKey(ctx, key)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := f(); err != nil {
			dg.mu.Lock()
			dg.errors = errors.Join(dg.errors, err)
			dg.mu.Unlock()
		}
		completed()
	}()
}

func withDoneGroup(ctx context.Context, cancelCause context.CancelCauseFunc, key any) context.Context {
	wg := &sync.WaitGroup{}
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		// Root doneGroup
		dg = &doneGroup{
			cancel:        cancelCause,
			cleanupGroups: []*sync.WaitGroup{wg},
		}
		return context.WithValue(ctx, key, dg)
	}
	// Add cleanupGroup to parent doneGroup
	dg.cleanupGroups = append(dg.cleanupGroups, wg)

	// Leaf doneGroup
	leafDg := &doneGroup{
		cancel:        cancelCause,
		cleanupGroups: []*sync.WaitGroup{wg},
	}
	return context.WithValue(ctx, key, leafDg)
}
