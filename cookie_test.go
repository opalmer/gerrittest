package gerrittest

import (
	"net/http"
	"net/url"

	. "gopkg.in/check.v1"
)

type CookieTest struct{}

var _ = Suite(&CookieTest{})

func (s *CookieTest) TestHostname(c *C) {
	expected := map[string]string{
		"127.0.0.1": "localhost",
		"localhost": "localhost",
		"foobar":    "foobar",
	}

	for name, value := range expected {
		c.Assert(hostname(&url.URL{Host: name}), Equals, value)
	}
}

func (s *CookieTest) TestCookieJar_SetCookies(c *C) {
	jar := NewCookieJar()
	u := &url.URL{Host: "127.0.0.1"}
	cookies := []*http.Cookie{{
		Name:  "foo",
		Path:  "/",
		Value: "hello",
	}}
	expected := map[string]map[string]*http.Cookie{}
	expected["localhost"] = map[string]*http.Cookie{}
	expected["localhost"]["foo"] = cookies[0]
	jar.SetCookies(u, cookies)
	c.Assert(jar.cookies, DeepEquals, expected)
}

func (s *CookieTest) TestCookieJar_Cookies(c *C) {
	jar := NewCookieJar()
	u := &url.URL{Host: "127.0.0.1"}
	cookies := []*http.Cookie{{
		Name:  "foo",
		Path:  "/",
		Value: "hello",
	}}

	jar.SetCookies(u, cookies)
	c.Assert(jar.Cookies(u), DeepEquals, cookies)
}
