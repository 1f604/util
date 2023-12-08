package util

import (
	"log"
	"net"
	"net/http"
	"time"
)

// Usage:
//
//	go web.RedirectToHTTPSFunc(":8080", 5 * time.Second, 5 * time.Second)
//
// Consider HSTS if your clients are browsers.
// I have manually tested this function and checked that it works whether your HTTP port is 80 or 8080.
// It will correct redirect to the correct https address. It will strip away any ports so you can't use a custom HTTPS port.
func RunRedirectHTTPToHTTPSServer(addr string, read_timeout time.Duration, write_timeout time.Duration, idle_timeout time.Duration, log_request bool) {
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  read_timeout,
		WriteTimeout: write_timeout,
		IdleTimeout:  idle_timeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if log_request {
				log.Print("-------------------------NEW HTTP REQUEST----------------------------")
				Nginx_Log_Received_Request("RunRedirectHTTPToHTTPSServer", req)
			}
			w.Header().Set("Connection", "close")
			host, _, err := net.SplitHostPort(req.Host) // Check if HTTP port is non-default
			if err != nil {                             // Assume port not in address
				host = req.Host
			}
			// log.Print("req.Host:", req.Host)
			// log.Print("req.URL:", req.URL.String())
			// log.Print("req.Host:", host)
			url := "https://" + host + req.RequestURI // assume TLS port is always 443 // DO NOT USE req.URL.String()!! Because that will append the scheme. So you will end up with
			// https://example.comhttp:/// if you set the URL Scheme to "http". SO DO NOT DO THAT. Use req.RequestURI instead.
			log.Print("Redirecting to HTTPS url: ", url)
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}
	log.Fatal(srv.ListenAndServe())
}
