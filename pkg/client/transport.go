package publichttp

import "net/http"

type AuthUserTransport struct {
	user string
	next http.RoundTripper
}

func NewAuthUserTransport(t http.RoundTripper, user string) http.RoundTripper {
	return &AuthUserTransport{user: user, next: t}
}

func (t *AuthUserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	const header = "X-Enapter-Auth-User"
	s := cloneRequest(req)
	s.Header.Set(header, t.user)
	return t.next.RoundTrip(s)
}

type AuthTokenTransport struct {
	token string
	next  http.RoundTripper
}

func NewAuthTokenTransport(t http.RoundTripper, token string) http.RoundTripper {
	return &AuthTokenTransport{token: token, next: t}
}

func (t *AuthTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	const header = "X-Enapter-Auth-Token"
	s := cloneRequest(req)
	s.Header.Set(header, t.token)
	return t.next.RoundTrip(s)
}

func cloneRequest(req *http.Request) *http.Request {
	shallow := new(http.Request)
	*shallow = *req
	for k, s := range req.Header {
		shallow.Header[k] = s
	}
	return shallow
}
