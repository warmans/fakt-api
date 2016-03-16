package middleware

import "net/http"

func AddCommonHeaders(handler http.Handler, headers map[string]string) http.Handler {
	return &CommonHeadersMiddleware{NextHandler: handler, Headers: headers}
}

type CommonHeadersMiddleware struct {
	NextHandler http.Handler
	Headers map[string]string
}

func (m *CommonHeadersMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	for k, v := range m.Headers {
		rw.Header().Add(k, v)
	}
	if r.Method == "OPTIONS" {
		rw.WriteHeader(204)
		return; //just send back the headers
	}
	m.NextHandler.ServeHTTP(rw, r)
}
