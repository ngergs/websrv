package server

import (
	"context"
	"net/http"
	"time"

	"github.com/ngergs/webserver/v2/utils"
	"github.com/rs/zerolog/log"
)

var sessionIddKey = &contextKey{val: "sessionId"}

const cookieName = "Session-Id"

type SessionCookieHandler struct {
	Next       http.Handler
	TimeToLife time.Duration
	Domain     string
	Storage    map[string]time.Time
}

func (handler *SessionCookieHandler) getNewSessionId() string {
	for {
		id := utils.GetRandomId(32)
		if _, ok := handler.Storage[id]; !ok {
			expires := time.Now().Add(handler.TimeToLife)
			handler.Storage[id] = expires
			return id
		}
	}
}

func (handler *SessionCookieHandler) readSessionIdCookie(r *http.Request) (sessionId string, ok bool) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == cookieName {
			return cookie.Value, true
		}
	}
	log.Ctx(r.Context()).Debug().Msgf("Cookie %s expired %d", cookieName)
	return "", false
}

func (handler *SessionCookieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "session-cookies")
	sessionId, ok := handler.readSessionIdCookie(r)
	if !ok {
		sessionId = handler.getNewSessionId()
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    sessionId,
			Domain:   handler.Domain,
			Path:     "/",
			MaxAge:   int(handler.TimeToLife.Seconds()),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
	ctx := context.WithValue(r.Context(), sessionIddKey, sessionId)
	handler.Next.ServeHTTP(w, r.WithContext(ctx))
}
