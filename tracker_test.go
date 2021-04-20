package alog

import (
	"context"
	"errors"
	"testing"
)

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
}
