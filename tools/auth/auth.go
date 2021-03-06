package auth

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"

	httpauth "github.com/abbot/go-http-auth"
	"github.com/gwaycc/mdoc/tools/cache"
	"github.com/gwaylib/errors"
)

const (
	REALM = "mdoc"
)

var (
	ErrNeedLogin = errors.New("Need login")
	ErrNeedPwd   = errors.New("Passwd failed")
	ErrReject    = errors.New("Too many login failures")
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

func UpdateAuthCache(username, token string) {
	authCache.Put(fmt.Sprintf(_AUTH_TOKEN_HEAD, username), token, 3600*24*_AUTH_EXPIRES_DAYS)
}

func GetAuthCache(username string) (string, bool) {
	cacheToken := authCache.Get(fmt.Sprintf(_AUTH_TOKEN_HEAD, username))
	if cacheToken == nil {
		return "", false
	}
	return cacheToken.(string), true
}
func DelAuthCache(username string) {
	authCache.Delete(fmt.Sprintf(_AUTH_TOKEN_HEAD, username))
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
	ips := []string{}
	addrInfo := strings.Split(r.RemoteAddr, ":")
	if len(addrInfo) > 0 {
		ips = append(ips, addrInfo[0])
	}

	for _, header := range r.Header["X-Forwarded-For"] {
		addrInfo := strings.Split(header, ":")
		if len(addrInfo) > 0 {
			ips = append(ips, addrInfo[0])
		}
	}
	return ips
}

func HashPasswd(user, realm, passwd string) string {
	return httpauth.H(fmt.Sprintf("%s:%s:%s", user, realm, passwd))
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
func (da *DigestAuth) CheckAuth(req *http.Request) (string, error) {
	username := ""
	auth := httpauth.DigestAuthParams(req.Header.Get(da.Headers.V().Authorization))
	if auth != nil {
		username = auth["username"]
	}
	if len(username) == 0 {
		return "", ErrNeedLogin.As("need username")
	}

	// detect whether it is an attack
	limitKey := fmt.Sprintf("%s_%+v", username, realIp(req))
	errTimes := getAuthLimit(limitKey)
	if errTimes > _AUTH_LIMIT_TIMES {
		return "", ErrReject.As(limitKey, errTimes)
	}

	// do login with password
	_, info := da.DigestAuth.CheckAuth(req)
	if info == nil {
		// auth failed
		updateAuthLimit(limitKey, errTimes+1)
		return "", ErrNeedPwd.As(auth)
	}

	// clean the errTimes when success
	updateAuthLimit(limitKey, 0)
	return username, nil
}

func readHttpResp(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	return string(body)
}
func AuthReq(server, uri, adminUser, adminPwd string, params url.Values) error {
	authResp, err := http.Head(server + uri)
	if err != nil {
		return errors.As(err)
	}
	defer authResp.Body.Close()
	if authResp.StatusCode != 401 {
		return errors.New(fmt.Sprintf("%d", authResp.StatusCode)).As(readHttpResp(authResp))
	}

	authParams := httpauth.DigestAuthParams(authResp.Header.Get("Www-Authenticate"))
	nonce := authParams["nonce"]
	opaque := authParams["opaque"]
	qop := authParams["qop"]
	//algorithm := authParams["algorithm"]
	realm := authParams["realm"]
	nc := fmt.Sprintf("%08d", rand.Intn(99999999))
	cnonce := fmt.Sprintf("%016x", rand.Intn(99999999))

	req, err := http.NewRequest("POST", server+uri, strings.NewReader(params.Encode()))
	if err != nil {
		return errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", algorithm=MD5, response="%s", opaque="%s", qop=%s, nc=%s, cnonce="%s"`,
		adminUser, realm, nonce, uri,
		httpauth.H(strings.Join([]string{
			httpauth.H(adminUser + ":" + realm + ":" + adminPwd),
			nonce, nc, cnonce, qop,
			httpauth.H("POST:" + uri),
		}, ":")),
		opaque, qop, nc, cnonce,
	))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.As(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("%d", resp.StatusCode)).As(readHttpResp(resp))
	}
	return nil
}
