package donegroup

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

var doneGroupKey = struct{}{}

type doneGroup struct {
	ctx    context.Context
	cancel context.CancelFunc

	// context.Context for after the context is canceled
	ctxw          context.Context
	cleanupGroups []*errgroup.Group
	mu            sync.Mutex
}

// WithCancel returns a copy of parent with a new Done channel and a doneGroup.
func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithCancelWithKey(ctx, doneGroupKey)
}

// WithCancelWithKey returns a copy of parent with a new Done channel and a doneGroup.
func WithCancelWithKey(ctx context.Context, key any) (context.Context, context.CancelFunc) {
	secondCtx, secondCancel := context.WithCancel(ctx)
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		dg = &doneGroup{
			ctx:    ctx,
			cancel: cancel,
		}
	}
	eg := new(errgroup.Group)
	dg.cleanupGroups = append(dg.cleanupGroups, eg)
	secondDg := &doneGroup{
		ctx:           dg.ctx,
		cancel:        dg.cancel,
		cleanupGroups: []*errgroup.Group{eg},
	}
	return context.WithValue(secondCtx, key, secondDg), secondCancel
}

// Clenup runs f when the context is canceled.
func Clenup(ctx context.Context, f func(ctx context.Context) error) error {
	return ClenupWithKey(ctx, doneGroupKey, f)
}

// ClenupWithKey runs f when the context is canceled.
func ClenupWithKey(ctx context.Context, key any, f func(ctx context.Context) error) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")
	}

	first := dg.cleanupGroups[0]
	first.Go(func() error {
		<-ctx.Done()
		<-dg.ctx.Done()
		dg.mu.Lock()
		ctx := dg.ctxw
		dg.mu.Unlock()
		return f(ctx)
	})
	return nil
}

// Wait blocks until the context is canceled.
func Wait(ctx context.Context) error {
	return WaitWithKey(ctx, doneGroupKey)
}

// Wait blocks until the context (ctx) is canceled. Then calls the function registered in Cleanup with context (ctxw).
func WaitWithContext(ctx, ctxw context.Context) error {
	return WaitWithContextAndKey(ctx, ctxw, doneGroupKey)
}

// WaitWithKey blocks until the context is canceled.
func WaitWithKey(ctx context.Context, key any) error {
	return WaitWithContextAndKey(ctx, context.Background(), key)
}

// WaitWithKeyAndContext blocks until the context is canceled. Then calls the function registered in Cleanup with context (ctxx).
func WaitWithContextAndKey(ctx, ctxw context.Context, key any) error {
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a doneGroup. Use donegroup.WithCancel to create a context with a doneGroup")
	}
	dg.mu.Lock()
	dg.ctxw = ctxw
	dg.mu.Unlock()
	<-ctx.Done()
	dg.cancel()
	eg, _ := errgroup.WithContext(ctxw)
	for _, g := range dg.cleanupGroups {
		eg.Go(g.Wait)
	}

	return eg.Wait()
}
