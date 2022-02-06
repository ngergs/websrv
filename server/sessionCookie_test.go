package server_test

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ngergs/webserver/server"
	"github.com/stretchr/testify/assert"
)

const cookieName = "testCookie"
const cookieLifeTime = time.Duration(10) * time.Second

func TestSessionCookieShouldBeAdded(t *testing.T) {
	// Setup test to get a session cookie
	w, r, next := getDefaultHandlerMocks()
	var responseHeader http.Header = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)
	handler := server.SessionCookieHandler(next, cookieName, cookieLifeTime)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)

	// check that cookie has been set and parse it
	responseCookie, ok := w.Header()["Set-Cookie"]
	assert.True(t, ok)
	cookie, sameSite := parseSetCookie(t, responseCookie[0])

	//static settings
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, "Strict", sameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, "", cookie.Domain)
	// dynamic settings
	assert.Equal(t, cookieName, cookie.Name)
	assert.Equal(t, int(cookieLifeTime.Seconds()), cookie.MaxAge)
	// allow some error here as this is set internally when the cookie is created
	expectedExpiresTime := time.Now().Add(cookieLifeTime)
	assert.True(t, cookie.Expires.After(expectedExpiresTime.Add(-time.Duration(1)*time.Second)))
	assert.True(t, cookie.Expires.Before(expectedExpiresTime.Add(time.Duration(1)*time.Second)))

	cookieValue := getCookieFromCtx(t, next.r.Context())
	assert.NotEqual(t, "", cookieValue)
}

func TestSessionCookieShouldNotAddedIfPresent(t *testing.T) {
	// Setup test to get a session cookie
	requestCookieValue := "test123"
	w, r, next := getDefaultHandlerMocks()
	var responseHeader http.Header = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)
	handler := server.SessionCookieHandler(next, cookieName, cookieLifeTime)
	r.Header.Set("Cookie", cookieName+"="+requestCookieValue)
	handler.ServeHTTP(w, r)

	//make sure that cookie has not been set in response
	_, ok := w.Header()["Set-Cookie"]
	assert.False(t, ok)

	cookieValue := getCookieFromCtx(t, next.r.Context())
	assert.Equal(t, requestCookieValue, cookieValue)
}

func getCookieFromCtx(t *testing.T, ctx context.Context) string {
	cookieVal := ctx.Value(server.SessionIdKey)
	assert.NotNil(t, cookieVal)
	return cookieVal.(string)
}

// parseSetCookie extracts a Cookie from the Set-Cookie header. The SameSite part is returned as a separate string, as the std lib http.readSetCookies method is private.
func parseSetCookie(t *testing.T, setCookie string) (cookie *http.Cookie, SameSite string) {
	cookieKeyValues := make(map[string]string)
	entries := strings.Split(setCookie, "; ")
	assert.Greater(t, len(entries), 0)
	name, value := splitSetCookieEntry(t, entries[0])

	for i := 1; i < len(entries); i++ {
		key, val := splitSetCookieEntry(t, entries[i])
		cookieKeyValues[key] = val
	}
	_, httpOnly := cookieKeyValues["HttpOnly"]
	_, secure := cookieKeyValues["Secure"]
	maxAge, err := strconv.Atoi(cookieKeyValues["Max-Age"])
	assert.Nil(t, err)
	expires, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", cookieKeyValues["Expires"])
	assert.Nil(t, err)
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cookieKeyValues["Path"],
		MaxAge:   maxAge,
		Expires:  expires,
		Secure:   secure,
		HttpOnly: httpOnly,
	}, cookieKeyValues["SameSite"]

}

func splitSetCookieEntry(t *testing.T, entry string) (key string, value string) {
	entryKeyVal := strings.Split(entry, "=")
	if len(entryKeyVal) != 2 {
		// for entries like HttpOnly or Secure
		return entryKeyVal[0], "true"
	}
	assert.Equal(t, 2, len(entryKeyVal))
	return entryKeyVal[0], entryKeyVal[1]
}
