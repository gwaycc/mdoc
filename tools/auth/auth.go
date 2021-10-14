package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	httpauth "github.com/abbot/go-http-auth"
	"github.com/google/uuid"
	"github.com/gwaycc/mdoc/tools/cache"
	"github.com/gwaylib/errors"
)

var (
	ErrNeedLogin = errors.New("Need login")
	ErrNeedPwd   = errors.New("Passwd failed")
)

// TODO: store to redis if need.
var (
	authCache = cache.NewMemoryCache(true)
)

const (
	_AUTH_TOKEN_HEAD   = "token_%s"
	_AUTH_LIMIT_HEAD   = "limit_%s"
	_AUTH_LIMIT_TIMES  = 4 // 4+1 times
	_AUTH_EXPIRES_DAYS = 7
)

func updateAuthToken(username, token string) {
	authCache.Put(fmt.Sprintf(_AUTH_TOKEN_HEAD, username), token, 3600*24*_AUTH_EXPIRES_DAYS)
}

func verifyAuthToken(username, token string) bool {
	cacheToken := authCache.Get(fmt.Sprintf(_AUTH_TOKEN_HEAD, username))
	if cacheToken == nil {
		return false
	}
	return cacheToken.(string) == token
}

func updateAuthLimit(key string, value int) {
	waitTime := int64(60 * 30)
	if value > _AUTH_LIMIT_TIMES {
		waitTime = int64(60 * 30 * (value - _AUTH_LIMIT_TIMES))
	}
	authCache.Put(fmt.Sprintf(_AUTH_LIMIT_HEAD, key), value, waitTime)
}

func getAuthLimit(key string) int {
	val := authCache.Get(fmt.Sprintf(_AUTH_LIMIT_HEAD, key))
	if val == nil {
		return 0
	}
	return val.(int)
}

func realIp(r *http.Request) []string {
	ips := []string{r.RemoteAddr}
	ips = append(ips, r.Header["X-Forwarded-For"]...)
	return ips
}

type DigestAuth struct {
	*httpauth.DigestAuth

	mutex sync.Mutex
}

// About SecretProvider
//
// plain text mode
// provider need return plain text, example: "hello".
//
// hash mode
// provider need return H(user + ":" + realm + ":" + passwd)
func NewDigestAuth(realm string, plainTextSecret bool, secret httpauth.SecretProvider) *DigestAuth {
	da := httpauth.NewDigestAuthenticator(realm, secret)
	da.PlainTextSecrets = plainTextSecret
	return &DigestAuth{
		DigestAuth: da,
	}
}

// rebuild httpauth.DigestAuth.CheckAuth
func (da *DigestAuth) CheckAuth(writer http.ResponseWriter, req *http.Request) (string, error) {
	username := ""

	// auth for loginedet/url
	loginCookie, _ := req.Cookie("login")
	if loginCookie != nil && loginCookie.Expires.Before(time.Now()) {
		val, err := url.ParseQuery(loginCookie.Value)
		if err == nil {
			// has login by token
			username = val.Get("username")
			token := val.Get("token")

			if verifyAuthToken(username, token) {
				// verify token pass
				updateAuthToken(username, token)
				return username, nil
			}
		}
	}
	auth := httpauth.DigestAuthParams(req.Header.Get(da.Headers.V().Authorization))
	if auth != nil {
		username = auth["username"]
	}
	if len(username) == 0 {
		da.RequireAuth(writer, req)
		return "", ErrNeedLogin.As("need username")
	}

	// detect whether it is an attack
	limitKey := fmt.Sprintf("%s_%+v", username, realIp(req))
	errTimes := getAuthLimit(limitKey)
	if errTimes > _AUTH_LIMIT_TIMES {
		writer.WriteHeader(403)
		writer.Write([]byte(fmt.Sprintf("Too many login failures: %d", errTimes)))
		return username, errors.New("Too many login failures").As(limitKey, errTimes)
	}

	// do login with password
	username, _ = da.DigestAuth.CheckAuth(req)
	if len(username) == 0 {
		// auth failed
		updateAuthLimit(limitKey, errTimes+1)
		da.RequireAuth(writer, req)
		return "", ErrNeedPwd.As(username)
	}

	// login success
	// write the token to the cookie
	loginToken := uuid.New().String()
	updateAuthToken(username, loginToken)
	cookiesVal := url.Values{
		"username": []string{username},
		"token":    []string{loginToken},
	}
	http.SetCookie(writer, &http.Cookie{
		Name:    "login",
		Value:   cookiesVal.Encode(),
		Expires: time.Now().AddDate(0, 0, _AUTH_EXPIRES_DAYS),
	})

	return username, nil
}
