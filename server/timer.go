package server

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var timerKey = &contextKey{val: "requestId"}

func TimerStartTOCtxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx = context.WithValue(ctx, timerKey, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logEnter(ctx context.Context, name string) {
	start := ctx.Value(timerKey)
	time.Now()
	if start != nil {
		log.Ctx(ctx).Debug().Msgf("entering %s: %v since request start", name, time.Since(start.(time.Time)))
	} else {
		log.Ctx(ctx).Debug().Msgf("entering %s", name)
	}
}
