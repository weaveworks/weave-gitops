package middleware

import (
	"context"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
)

var ctxKey = new(int)

func format(svc, meth, status string, t time.Time, err error) *log.Entry {
	m := log.WithFields(log.Fields{
		"service":  svc,
		"method":   meth,
		"status":   status,
		"duration": time.Since(t),
	})

	if err != nil {
		m = m.WithFields(log.Fields{"error": err.Error()})
	}

	return m
}

func getRequestParams(ctx context.Context) (svc string, meth string, status string, code int, t time.Time) {
	svc, ok := twirp.ServiceName(ctx)
	meth, ok = twirp.MethodName(ctx)
	status, ok = twirp.StatusCode(ctx)

	if !ok {
		log.Error("could not get svc or method name")
		return
	}

	code, err := strconv.Atoi(status)

	if err != nil {
		log.Error("could not get status code")
		code = 0
	}

	t, ok = ctx.Value(ctxKey).(time.Time)

	if !ok {
		log.Error("could not get request duration")
		t = time.Time{}
	}

	return svc, meth, status, code, t
}

func LoggingHooks() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			startTime := time.Now()
			ctx = context.WithValue(ctx, ctxKey, startTime)
			return ctx, nil
		},
		ResponseSent: func(ctx context.Context) {
			svc, meth, status, code, t := getRequestParams(ctx)

			msg := format(svc, meth, status, t, nil)

			if code < 400 {
				msg.Debug("request success")
			}

		},
		Error: func(ctx context.Context, e twirp.Error) context.Context {
			svc, meth, status, code, t := getRequestParams(ctx)

			msg := format(svc, meth, status, t, e)

			if code > 399 && code < 499 {
				msg.Warn("request error")
			}

			if code > 499 {
				msg.Error("internal error")
			}

			return ctx
		},
	}
}
