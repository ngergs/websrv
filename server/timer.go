package server

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var TimerKey = &ContextKey{val: "requestId"}

func TimerStartTOCtxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, TimerKey, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logEnter(ctx context.Context, name string) {
	start := ctx.Value(TimerKey)
	if start != nil {
		log.Ctx(ctx).Debug().Msgf("entering %s: %v since request start", name, time.Since(start.(time.Time)))
	} else {
		log.Ctx(ctx).Debug().Msgf("entering %s", name)
	}
}
