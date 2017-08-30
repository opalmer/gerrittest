package gerrittest

import (
	"net/http"
	"net/url"
	"sync"
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
	cookies map[string]map[string]*http.Cookie
}

// SetCookies will set cookies for the given url.
func (c *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	host := hostname(u)
	hostCookies, ok := c.cookies[host]
	if !ok {
		hostCookies = map[string]*http.Cookie{}
		c.cookies[host] = hostCookies
	}

	keys := []string{}
	for _, cookie := range cookies {
		hostCookies[cookie.Name] = cookie
		keys = append(keys, cookie.Name)
	}
}

// Cookies returns the cookies associated with the given url.
func (c *CookieJar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	host := hostname(u)
	names := []string{}
	for _, cookie := range c.cookies[host] {
		names = append(names, cookie.Name)
		cookies = append(cookies, cookie)
	}
	return cookies
}

// NewCookieJar constructs and returns a *CookieJar struct.
func NewCookieJar() *CookieJar {
	return &CookieJar{
		mtx:     &sync.Mutex{},
		cookies: map[string]map[string]*http.Cookie{},
	}
}
