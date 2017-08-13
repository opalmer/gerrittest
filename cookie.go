package gerrittest

import (
	"net/http"
	"net/url"
	"sync"

	log "github.com/Sirupsen/logrus"
)

func hostname(u *url.URL) string {
	hostname := u.Hostname()
	switch u.Hostname() {
	case "127.0.0.1", "localhost":
		return "localhost"
	}
	return hostname
}

// CookieJar is an implementation of a cookie jar similar to
// http/cookiejar. This implementation is only intended for
// local development.
type CookieJar struct {
	mtx     *sync.Mutex
	log     *log.Entry
	cookies map[string]map[string]*http.Cookie
}

// SetCookies will set cookies for the given url.
func (c *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	logger := c.log.WithFields(log.Fields{
		"action": "set",
	})
	host := hostname(u)
	hostCookies, ok := c.cookies[host]
	if !ok {
		logger = logger.WithField("init", true)
		hostCookies = map[string]*http.Cookie{}
		c.cookies[host] = hostCookies
	}

	keys := []string{}
	for _, cookie := range cookies {
		hostCookies[cookie.Name] = cookie
		keys = append(keys, cookie.Name)
	}
	logger.WithFields(log.Fields{
		"set":     keys,
		"cookies": len(hostCookies),
	}).Debug()
}

// Cookies returns the cookies associated with the given url.
func (c *CookieJar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	logger := c.log.WithFields(log.Fields{
		"action": "get",
	})
	host := hostname(u)
	hostCookies, _ := c.cookies[host]
	names := []string{}
	for _, cookie := range hostCookies {
		names = append(names, cookie.Name)
		cookies = append(cookies, cookie)
	}
	logger.WithFields(log.Fields{
		"names": names,
	}).Debug()
	return cookies
}

// NewCookieJar constructs and returns a *CookieJar struct.
func NewCookieJar() *CookieJar {
	return &CookieJar{
		mtx:     &sync.Mutex{},
		cookies: map[string]map[string]*http.Cookie{},
		log:     log.WithField("cmp", "cookie"),
	}
}
