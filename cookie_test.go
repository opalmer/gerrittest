package gerrittest

import (
	"net/http"
	"net/url"
	"testing"
)

func TestHostname(t *testing.T) {
	expected := map[string]string{
		"127.0.0.1": "localhost",
		"localhost": "localhost",
		"foobar":    "foobar",
	}
	for name, value := range expected {
		if hostname(&url.URL{Host: name}) != value {
			t.Fatalf("%s != %s", name, value)
		}
	}
}

func TestCookieJar_SetCookies(t *testing.T) {
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
	if jar.cookies["localhost"]["foo"] != cookies[0] {
		t.Fatal()
	}
}

func TestCookieJar_Cookies(t *testing.T) {
	jar := NewCookieJar()
	u := &url.URL{Host: "127.0.0.1"}
	cookies := []*http.Cookie{{
		Name:  "foo",
		Path:  "/",
		Value: "hello",
	}}

	jar.SetCookies(u, cookies)
	result := jar.Cookies(u)

	for i, cookie := range cookies {
		stored := result[i]
		if cookie.Name != stored.Name {
			t.Fatal()
		}
		if cookie.Path != stored.Path {
			t.Fatal()
		}
		if cookie.Value != stored.Value {
			t.Fatal()
		}
	}
}
