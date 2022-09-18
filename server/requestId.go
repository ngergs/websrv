package server

import (
	"context"
	"github.com/ngergs/websrv/internal/random"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ContextKey is a struct used for storing relevant keys in the request context.
type ContextKey struct {
	val string
}

// RequestIdKey is the context key used for storing the request id in the request context.
var RequestIdKey = &ContextKey{val: "requestId"}

// RequestIdToCtxHandler generates a random request-id and adds it to the request context under the RequestIdKey.
func RequestIdToCtxHandler(next http.Handler) http.Handler {
	randGen := random.NewBufferedRandomIdGenerator(32, 16)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "request-id")
		ctx := r.Context()
		requestId := randGen.GetRandomId()
		ctx = log.With().Str("requestId", requestId).Logger().WithContext(ctx)
		ctx = context.WithValue(ctx, RequestIdKey, requestId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
