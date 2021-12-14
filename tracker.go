package alog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	KeyTracker            = "_alog_Tracker"
	KeyTrackID            = "_alog_TrackID"
	GinHttpResponseHeader = "X-Track"
)

type V map[string]interface{}

type Tracker struct {
	ID         string
	Exceptions []*Exception
	Request    *sentry.Request
	lock       sync.Mutex
}
type Exception struct {
	Error  error
	Stacks []*ExceptionStack
}
type ExceptionStack struct {
	Filename string
	Package  string
	Function string
	Lineno   int
	Vars     V
}

func NewTracker() *Tracker {
	return &Tracker{
		ID: string(RandomBytes(8)),
	}
}

func (t *Tracker) Print() {
	var buf = time.Now().UTC().AppendFormat([]byte("[TRACK] "), "2006/01/02 15:04:05 [")
	buf = append(buf, []byte(t.ID)...)
	buf = append(buf, ']', '\n')
	for i, ex := range t.Exceptions {
		buf = append(buf, []byte(fmt.Sprintf("  [%d] %s\n", i, ex.Error.Error()))...)
		for _, stack := range ex.Stacks {
			js, _ := json.Marshal(stack.Vars)
			buf = append(buf, fmt.Sprintf("  %s:%d: %s\n", stack.Filename, stack.Lineno, js)...)
		}
	}
	os.Stdout.Write(buf)
}

func GetTrackID(ctx context.Context) string {
	tracker := GetTracker(ctx)
	if tracker != nil {
		return tracker.ID
	}
	return ""
}

func GetTracker(ctx context.Context) *Tracker {
	tracker, ok := ctx.Value(KeyTracker).(*Tracker)
	if ok {
		return tracker
	}
	return nil
}

func setValuesToStack(stack *ExceptionStack, trackValues []map[string]interface{}) {
	// 合并传入的参数
	for _, tv := range trackValues {
		if tv == nil {
			continue
		}
		if stack.Vars == nil {
			stack.Vars = tv
		} else {
			for k, v := range tv {
				stack.Vars[k] = v
			}
		}
	}
}

func ce(ctx context.Context, err error, trackValues []map[string]interface{}) {
	var stack ExceptionStack
	setValuesToStack(&stack, trackValues)

	// 追踪当前的栈信息
	if pc, file, line, ok := runtime.Caller(2); ok {
		stack.Filename = file
		stack.Lineno = line
		f := runtime.FuncForPC(pc)
		words := strings.Split(f.Name(), ".")
		if len(words) == 2 {
			stack.Package, stack.Function = words[0], words[1]
		}
	} else {
		return
	}

	// 如果ctx里有Tracker就放进去统一处理，否则直接丢到日志里去。
	tracker := GetTracker(ctx)
	if tracker != nil {
		tracker.lock.Lock()
		defer tracker.lock.Unlock()
		for _, ex := range tracker.Exceptions {
			if errors.Is(ex.Error, err) {
				ex.Stacks = append(ex.Stacks, &stack)
				return
			}
		}
		tracker.Exceptions = append(tracker.Exceptions, &Exception{
			Error:  err,
			Stacks: []*ExceptionStack{&stack},
		})
	} else {
		js, _ := json.Marshal(stack.Vars)
		ERROR.Printf("%s:%d: %s. %s\n", stack.Filename, stack.Lineno, err, js)
	}
}

// CE 意思是 CheckError ，为了方便按键而起这个名字。
func CE(ctx context.Context, err error, trackValues ...map[string]interface{}) {
	if err == nil {
		return
	}
	ce(ctx, err, trackValues)
}

// CEI 意思是 Check Error Interface，可以灵活处理interface{}。
// 建议不要用在 recover() 的情况，会丢失 panic() 的位置，请使用 CERecover 替代。
// 建议不要使用可比较值，否则可能与其他错误栈混在一起，最好用指针或者接口变量。
func CEI(ctx context.Context, err interface{}, trackValues ...map[string]interface{}) {
	if err == nil {
		return
	}
	e, ok := err.(error)
	if ok {
		ce(ctx, e, trackValues)
	} else {
		ce(ctx, errors.New(fmt.Sprintf("Interface<%T>: %v", err, err)), trackValues)
	}
}

func BuildSentryEvent(tracker *Tracker) *sentry.Event {
	if tracker.Exceptions == nil {
		return nil
	}

	var exceptions []sentry.Exception
	for _, ex := range tracker.Exceptions {
		var trace sentry.Stacktrace
		for i := len(ex.Stacks) - 1; i >= 0; i-- {
			stack := ex.Stacks[i]
			trace.Frames = append(trace.Frames, sentry.Frame{
				Filename: stack.Filename,
				Function: stack.Function,
				Package:  stack.Package,
				Lineno:   stack.Lineno,
				Vars:     stack.Vars,
			})
		}
		var exception = sentry.Exception{
			Value:      ex.Error.Error(),
			Type:       reflect.TypeOf(ex.Error).String(),
			Stacktrace: &trace,
		}
		exceptions = append(exceptions, exception)
	}

	event := sentry.Event{
		Extra:     map[string]interface{}{"TrackID": tracker.ID},
		Level:     "error",
		Timestamp: time.Now().UTC(),
		Sdk: sentry.SdkInfo{
			Name:    "sentry.go",
			Version: sentry.Version,
		},
		Release:   ConfigAppVersion,
		Exception: exceptions,
		Request:   tracker.Request,
	}

	return &event
}

func SendSentryEvent(event *sentry.Event) error {
	if event == nil {
		return nil
	}
	js, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", ConfigSentryUrl, bytes.NewReader(js))
	auth := fmt.Sprintf("Sentry sentry_key=%s", ConfigSentryPublicKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sentry-Auth", auth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Sentry上报异常: %s. %s", resp.Status, body))
	}
	return nil
}

func BuildAndSendSentryEvent(tracker *Tracker) {
	event := BuildSentryEvent(tracker)
	err := SendSentryEvent(event)
	if err != nil {
		ERROR.Println(err)
	}
}

func CheckTracker(tracker *Tracker) {
	if tracker == nil {
		return
	}
	if tracker.Exceptions != nil {
		if ConfigSentryUrl == "" {
			tracker.Print()
		} else {
			go BuildAndSendSentryEvent(tracker)
		}
	}
}
