package alog

import (
	"context"
	"errors"
	"testing"
)

func TestWithTracker(t *testing.T) {
	throw := func() error {
		return errors.New("记录这个错误:" + string(RandomBytes(4)))
	}

	t.Run("常规记录错误", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		CE(ctx, throw(), V{"一些变量": "一些值"})
	})
	t.Run("常规记录错误(多级)", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		if err := func() error {
			e := throw()
			if e != nil {
				CE(ctx, e, V{"一些变量": "一些值"})
			}
			return e
		}(); err != nil {
			CE(ctx, err, V{"上层的变量": "上层的值"})
		}
	})
	t.Run("常规记录错误(多级多个)", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		if err := func() error {
			e := throw()
			if e != nil {
				CE(ctx, e, V{"一些变量": "一些值"})
			}
			return e
		}(); err != nil {
			CE(ctx, err, V{"上层的变量1": "上层的值"})
		}
		if err := throw(); err != nil {
			CE(ctx, err, V{"上层的变量2": "这里又发生错误了"})
		}
	})
	t.Run("无事发生", func(t *testing.T) {
		_, cancel := WithTracker(context.Background())
		defer cancel()
	})
}
