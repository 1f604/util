package util

import (
	web_types "fileserver/pkg/util/web_types"
	"log"
	"net/http"
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

func RedirectWrapper(handler http.Handler, scheme string, url_redirect_map web_types.URLRedirectMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("------------------------REDIRECT WRAPPER: NEW HTTP REQUEST----------------------------")
		Nginx_Log_Received_Request(r)
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
