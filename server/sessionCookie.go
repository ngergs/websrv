package server

import (
	"context"
	"github.com/ngergs/websrv/internal/random"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// SessionIdKey is the ContextKey under which the current sessionId can be found
var SessionIdKey = &ContextKey{val: "sessionId"}

func readSessionIdCookie(r *http.Request, cookieName string) (sessionId string, ok bool) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == cookieName {
			return cookie.Value, true
		}
	}
	log.Ctx(r.Context()).Debug().Msgf("Cookie %s not present in request", cookieName)
	return "", false
}

// SessionCookieHandler reads the cookieName cookie from the request and adds if to the context unter the SessionIdKey if present.
// If absent it generates a new sessionId and adds it to the context and the HTTP Set-Cookie Response header.
func SessionCookieHandler(next http.Handler, cookieName string, cookieTimeToLife time.Duration) http.Handler {
	randGen := random.NewBufferedRandomIdGenerator(32, 16)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), cookieName)
		sessionId, ok := readSessionIdCookie(r, cookieName)
		if !ok {
			// random collisions are not problematic for CSP nonces, so we can just take what we get
			sessionId = randGen.GetRandomId()
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    sessionId,
				Path:     "/",
				MaxAge:   int(cookieTimeToLife.Seconds()),
				Expires:  time.Now().Add(cookieTimeToLife),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}
		ctx := context.WithValue(r.Context(), SessionIdKey, sessionId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
