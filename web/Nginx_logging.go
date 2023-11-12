package util

import (
	"log"
	"net/http"
)

func Nginx_Log_Received_Request(r *http.Request) {
	// log format is based on default nginx log format
	/*
	  log_format combined '$remote_addr - $remote_user [$time_local] '
	                      '"$request" $body_bytes_sent '
	                      '"$http_referer" "$http_user_agent"';
	*/
	var http_or_https string
	if r.TLS == nil {
		http_or_https = "HTTP"
	} else {
		http_or_https = "HTTPS"
	}
	log.Printf("RemoteAddr: %s Method: %s Scheme: %s Protocol: %s Host: %#q URL: %#q RequestURI:%#q ContentLength:%d Referrer:%#q UserAgent:%#q", r.RemoteAddr, r.Method, http_or_https, r.Proto, r.Host, r.URL, r.RequestURI, r.ContentLength, r.Referer(), r.UserAgent())
}
