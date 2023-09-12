package job

import (
	"context"

	"github.com/runabol/tork"
)

type EventType string

const (
	// StateChange occurs when a job's state changes.
	// Handler can inspect the job's State property
	// in order to determine what state the job is at.
	StateChange = "STATE_CHANGE"
	// Read occurs when a Job is read by the client
	// through the API.
	Read = "READ"
)

type HandlerFunc func(ctx context.Context, et EventType, j *tork.Job) error

func NoOpHandlerFunc(ctx context.Context, et EventType, j *tork.Job) error { return nil }

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

func ApplyMiddleware(h HandlerFunc, mws []MiddlewareFunc) HandlerFunc {
	return func(ctx context.Context, et EventType, t *tork.Job) error {
		nx := next(ctx, 0, mws, h)
		return nx(ctx, et, t)
	}
}

func next(ctx context.Context, index int, mws []MiddlewareFunc, h HandlerFunc) HandlerFunc {
	if index >= len(mws) {
		return h
	}
	return mws[index](next(ctx, index+1, mws, h))
}
