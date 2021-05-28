package middleware

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twitchtv/twirp"
)

var durationKey = new(int)

func MetricsHooks(durations *prometheus.HistogramVec) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			startTime := time.Now()
			ctx = context.WithValue(ctx, durationKey, startTime)
			return ctx, nil
		},

		ResponseSent: func(ctx context.Context) {
			startTime, ok := ctx.Value(durationKey).(time.Time)

			if !ok {
				return
			}

			svc, _ := twirp.ServiceName(ctx)
			method, _ := twirp.MethodName(ctx)
			status, _ := twirp.StatusCode(ctx)
			durations.WithLabelValues(svc, method, status).Observe(time.Since(startTime).Seconds())
		},
	}
}
