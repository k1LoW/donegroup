package donegroup

import (
	"context"
	"testing"
	"time"
)

func TestDoneGroup(t *testing.T) {
	ctx, cancel := WithCancel(context.Background())

	cleanup := false

	if err := Clenup(ctx, func() error {
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

func TestMultiCleanup(t *testing.T) {
	ctx, cancel := WithCancel(context.Background())

	cleanup := 0

	for i := 0; i < 10; i++ {
		if err := Clenup(ctx, func() error {
			time.Sleep(10 * time.Millisecond)
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
	firstCtx, firstCancel := WithCancel(context.Background())
	secondCtx, secondCancel := WithCancel(firstCtx)

	firstCleanup := 0
	secondCleanup := 0

	for i := 0; i < 10; i++ {
		if err := Clenup(firstCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			firstCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Clenup(secondCtx, func() error {
			time.Sleep(10 * time.Millisecond)
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
	rootCtx, rootCancel := WithCancel(context.Background())
	leafCtx, _ := WithCancel(rootCtx)

	rootCleanup := 0
	leafCleanup := 0

	for i := 0; i < 10; i++ {
		if err := Clenup(rootCtx, func() error {
			time.Sleep(10 * time.Millisecond)
			rootCleanup += 1
			return nil
		}); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		if err := Clenup(leafCtx, func() error {
			time.Sleep(10 * time.Millisecond)
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
