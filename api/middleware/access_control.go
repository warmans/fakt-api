package middleware

import "net/http"

func AddCommonHeaders(nextHandler http.Handler, headers map[string]string) http.Handler {
	return &CommonHeadersMiddleware{next: nextHandler, headers: headers}
}

type CommonHeadersMiddleware struct {
	next    http.Handler
	headers map[string]string
}

func (m *CommonHeadersMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	for k, v := range m.headers {
		rw.Header().Add(k, v)
	}
	if r.Method == "OPTIONS" {
		rw.WriteHeader(204)
		return; //just send back the headers
	}
	m.next.ServeHTTP(rw, r)
}
