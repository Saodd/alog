package alog

import (
	"context"
	"errors"
	"testing"
)

type UncomparableObject map[string]string

func (e UncomparableObject) Error() string {
	return ""
}

func TestCE(t *testing.T) {
	t.Run("没有Tracker时直接打出来！", func(t *testing.T) {
		CE(context.Background(), errors.New("要输出:null"), nil, nil)
	})
	t.Run("多个nil的处理！", func(t *testing.T) {
		CE(context.Background(), errors.New("要输出:null"), nil, nil)
	})
	t.Run("多个map的合并", func(t *testing.T) {
		CE(context.Background(), errors.New(`要输出:{"a":1,"b":2}`), V{"a": 1}, V{"b": 2})
	})
	t.Run("多个map的覆盖", func(t *testing.T) {
		CE(context.Background(), errors.New(`要输出:{"a":2,"b":2}`), V{"a": 1}, V{"a": 2, "b": 2})
	})
	t.Run("多个map和bil的混合覆盖", func(t *testing.T) {
		CE(context.Background(), errors.New(`要输出:{"a":2,"b":2}`), nil, V{"a": 1}, nil, V{"a": 2, "b": 2}, nil)
	})
	t.Run("兼容不可比较的错误值，不能panic，应该作为两个错误", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		err := UncomparableObject{}
		CE(ctx, err)
		CE(ctx, err)
	})
	t.Run("错误值的比较，应该显示为同一个", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		err := &UncomparableObject{}
		CE(ctx, err)
		CE(ctx, err)
	})
}

func TestCEI(t *testing.T) {
	t.Run("panic nil 应该什么都不输出", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer func() {
			CEI(ctx, recover())
		}()
		panic(nil)
	})
	t.Run("panic error 应该看见一个追踪信息", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer func() {
			CEI(ctx, recover())
		}()
		panic(errors.New("看见我是错误！"))
	})
	t.Run("panic any 应该看见一个追踪信息", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer func() {
			CEI(ctx, recover())
		}()
		panic("看见我！")
	})
	t.Run("panic nil interface 应该看见一个追踪信息", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer func() {
			CEI(ctx, recover())
		}()
		type A struct{}
		type AI interface{}
		var a AI = (*A)(nil)
		panic(a) // 目前认为，带类型的nil是应当作为错误来处理的，不该忽略
	})
}

func TestCERecover(t *testing.T) {
	t.Run("panic error 看见一串追踪栈", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer CERecover(ctx, V{"data": "只看见我一次"})
		panic(errors.New("看见我是错误！"))
	})
	t.Run("panic error 看见一串追踪栈", func(t *testing.T) {
		ctx, cancel := WithTracker(context.Background())
		defer cancel()
		defer CERecover(ctx, V{"data": "只看见我一次"})
		var a []int
		a[0] = 0
	})
}
