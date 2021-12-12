package alog

import (
	"context"
	"testing"
)

func TestTraceStack(t *testing.T) {
	t.Run("不仅不能抛出异常，还要打出日志哦！", func(t *testing.T) {
		defer Recover(context.Background())
		panic("这个崩溃原因也要打在日志里哦！")
	})
	t.Run("来自运行时的异常", func(t *testing.T) {
		defer Recover(context.Background())
		var a []int
		a[0] = 0
	})
	t.Run("没有异常时也要正常工作！", func(t *testing.T) {
		defer Recover(context.Background())
	})
	t.Run("有Tracker！", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer Recover(ctx)
		panic("这个崩溃原因也要打在日志里哦！")
	})
	t.Run("有Trcker没有异常时也要正常工作！", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer Recover(ctx)
	})
}

func TestRecoverError(t *testing.T) {
	t.Run("Recover的异常要返回回去！", func(t *testing.T) {
		f := func() (err error) {
			ctx, cancel := WithTracker(context.Background())
			defer cancel()
			defer CERecoverError(ctx, &err)
			panic("故意的")
		}
		if err := f(); err == nil {
			t.Error("没有正确返回异常", err)
		}
	})
	t.Run("无事发生", func(t *testing.T) {
		f := func() (err error) {
			ctx, cancel := WithTracker(context.Background())
			defer cancel()
			defer CERecoverError(ctx, &err)
			return nil
		}
		if err := f(); err != nil {
			t.Error("应该无事发生", err)
		}
	})
}
