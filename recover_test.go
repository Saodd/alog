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
