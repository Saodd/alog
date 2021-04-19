package alog

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// GinWithLogger returns gin.HandlerFunc, it should be used as a middleware.
// Compare to gin.Logger(), it prints "TrackerID" for each request in addition.
func GinWithLogger() gin.HandlerFunc {
	formatter := func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			// Truncate in a golang < 1.8 safe way
			param.Latency = param.Latency - param.Latency%time.Second
		}
		var trackID, ok = param.Keys[KeyTrackID].(string)
		if ok {
			return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s[%-8s] %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				trackID,
				param.Path,
				param.ErrorMessage,
			)
		}
		return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)

	}

	cfg := gin.LoggerConfig{
		Formatter: formatter,
	}
	return gin.LoggerWithConfig(cfg)
}

// GinWithTracker returns gin.HandlerFunc, it should be used as a middleware.
// It injects a Tracker to *gin.Context for each request.
func GinWithTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracker := NewTracker()
		c.Set(KeyTracker, tracker)
		c.Set(KeyTrackID, tracker.ID)
		c.Header(GinHttpResponseHeader, tracker.ID)

		c.Next()

		if tracker.Exceptions != nil {
			tracker.Request = sentry.NewRequest(c.Request)
			if ConfigSentryUrl == "" {
				go tracker.Print()
			} else {
				go BuildAndSendSentryEvent(tracker)
			}
		}
	}
}

// GinWithRecover returns gin.HandlerFunc, it should be used as a middleware.
// It recovers your server from panic, and record the error in Tracker (to handle it later).
func GinWithRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a condition that warrants a panic stack trace.
				// 逻辑来自gin.RecoveryWithWriter()
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				TraceStack(ctx, err)

				// If the connection is dead, we can't write a status to it.
				// 逻辑来自gin.RecoveryWithWriter()
				if brokenPipe {
					ctx.Error(err.(error))
					ctx.Abort()
				} else {
					ctx.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		ctx.Next()
	}
}
