package util

import (
	"log"
	"net/http"

	web_types "github.com/1f604/util/web_types"
)

/*
Complete pattern for constructing middleware: https://www.alexedwards.net/blog/making-and-using-middleware

func exampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

*/

func RedirectWrapper(handler http.Handler, scheme string, url_redirect_map web_types.URLRedirectMap, log_request bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if log_request {
			log.Print("------------------------REDIRECT WRAPPER: NEW HTTP REQUEST----------------------------")
			Nginx_Log_Received_Request("RedirectWrapper", r)
		}
		matched, ok := (*url_redirect_map.Map)[r.URL.Path]
		// If the key exists
		if ok {
			dest_url := scheme + r.Host + matched // DO NOT USE req.URL.String()!! Because that will append the scheme. So you will end up with
			log.Print("Redirect wrapper: URL found in redirect rule map. Redirecting to url: ", dest_url)
			http.Redirect(w, r, dest_url, http.StatusMovedPermanently)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
