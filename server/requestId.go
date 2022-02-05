package server

import (
	"context"
	"net/http"

	"github.com/ngergs/webserver/v2/utils"
	"github.com/rs/zerolog/log"
)

type ContextKey struct {
	val string
}

var RequestIdKey = &ContextKey{val: "requestId"}

func RequestIdToCtxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "request-id")
		ctx := r.Context()
		requestId := utils.GetRandomId(32)
		log := log.With().Str("requestId", requestId).Logger()
		ctx = log.WithContext(ctx)
		ctx = context.WithValue(ctx, RequestIdKey, requestId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
