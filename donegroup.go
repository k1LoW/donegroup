package donegroup

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var doneGroupKey = struct{}{}

type doneGroup struct {
	ctx           context.Context
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
		dg = &doneGroup{}
	}
	eg := new(errgroup.Group)
	dg.cleanupGroups = append(dg.cleanupGroups, eg)
	secondDg := &doneGroup{cleanupGroups: []*errgroup.Group{eg}}
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
		return errors.New("donegroup: context does not contain a donegroup. Use donegroup.WithCancel to create a context with a donegroup")
	}

	first := dg.cleanupGroups[0]
	first.Go(func() error {
		<-ctx.Done()
		return dg.goWithCtx(f)
	})
	return nil
}

// Wait blocks until the context is canceled.
func Wait(ctx context.Context) error {
	return WaitWithKey(ctx, doneGroupKey)
}

// Wait blocks until the context is canceled or the timeout is reached.
func WaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	return WaitWithKeyAndTimeout(ctx, doneGroupKey, timeout)
}

// WaitWithKey blocks until the context is canceled.
func WaitWithKey(ctx context.Context, key any) error {
	return WaitWithKeyAndTimeout(ctx, key, 0)
}

// WaitWithKeyAndTimeout blocks until the context is canceled or the timeout is reached.
func WaitWithKeyAndTimeout(ctx context.Context, key any, timeout time.Duration) error {
	<-ctx.Done()
	dg, ok := ctx.Value(key).(*doneGroup)
	if !ok {
		return errors.New("donegroup: context does not contain a donegroup. Use donegroup.WithCancel to create a context with a donegroup")
	}
	ctxx := context.Background()
	var cancel context.CancelFunc
	if timeout != 0 {
		ctxx, cancel = context.WithTimeout(ctxx, timeout)
		defer cancel()
	}
	dg.mu.Lock()
	dg.ctx = ctxx
	dg.mu.Unlock()
	eg, _ := errgroup.WithContext(ctxx)
	for _, g := range dg.cleanupGroups {
		eg.Go(g.Wait)
	}

	return eg.Wait()
}

func (dg *doneGroup) goWithCtx(f func(ctx context.Context) error) error {
	dg.mu.Lock()
	ctx := dg.ctx
	dg.mu.Unlock()
	return f(ctx)
}
