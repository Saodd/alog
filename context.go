package alog

import (
	"context"
)

func CheckContext(ctx context.Context) {
	CheckTracker(GetTracker(ctx))
}

func WithTracker(parent context.Context) (context.Context, context.CancelFunc) {
	tracker := NewTracker()
	ctx := context.WithValue(parent, KeyTracker, tracker)
	return ctx, func() {
		CheckTracker(tracker)
	}
}
