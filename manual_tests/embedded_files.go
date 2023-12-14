package main

import (
	"embed"
	"net/http"
	"net/url"

	"github.com/1f604/util"
	web "github.com/1f604/util/web"
)

//go:embed all:static
var static_content embed.FS

func main() {
	_, err := static_content.Open("static/hello.txt")
	util.Check_err(err)
	var w http.ResponseWriter
	var r *http.Request = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/static/a.a/content",
		},
	}

	web.SafelyServeFileEmbedded(w, r, "/static/", "static/", static_content, false)
}
