package server

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type contextKey string

const requestIdKey contextKey = "requestId"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func getRandomId(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RequestIdToCtxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "request-id")
		ctx := r.Context()
		requestId := getRandomId(32)
		log := log.With().Str("requestId", requestId).Logger()
		ctx = log.WithContext(ctx)
		ctx = context.WithValue(ctx, requestIdKey, requestId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
