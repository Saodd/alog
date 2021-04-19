package alog

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

func TraceStack(ctx context.Context, e interface{}) {
	err, ok := e.(error)
	if !ok {
		err = errors.New(fmt.Sprint(e))
	}

	// 取出所有栈
	var stacks []*ExceptionStack
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		var stack ExceptionStack
		stack.Filename = file
		stack.Lineno = line
		f := runtime.FuncForPC(pc)
		words := strings.Split(f.Name(), ".")
		if len(words) == 2 {
			stack.Package, stack.Function = words[0], words[1]
		}
		stacks = append(stacks, &stack)
	}

	// 如果ctx里有Tracker就放进去统一处理，否则直接丢到日志里去。
	tracker := GetTracker(ctx)
	if tracker != nil {
		tracker.lock.Lock()
		defer tracker.lock.Unlock()
		for _, ex := range tracker.Exceptions {
			if ex.Error == err {
				ex.Stacks = append(ex.Stacks, stacks...)
				return
			}
		}
		tracker.Exceptions = append(tracker.Exceptions, &Exception{
			Error:  err,
			Stacks: stacks,
		})
	} else {
		b := strings.Builder{}
		b.WriteString(err.Error())
		b.WriteByte('\n')
		for _, stack := range stacks {
			b.WriteString(fmt.Sprintf("    %s:%d\n", stack.Filename, stack.Lineno))
		}
		RECOVER.Println(b.String())
	}
}

func Recover(ctx context.Context) {
	if err := recover(); err != nil {
		TraceStack(ctx, err)
	}
}
