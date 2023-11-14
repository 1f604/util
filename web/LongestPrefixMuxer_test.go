package util_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	web "github.com/1f604/util/web"
)

func handle_hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!")) //nolint
}

func handle_helloworld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!")) //nolint
}

func handle_helloworldagain(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world again!")) //nolint
}

func fallback_handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("fallback!")) //nolint
}

func Test_LongestPrefixMuxer(t *testing.T) {
	t.Parallel()

	mux_entries := []*web.MuxEntry{
		web.NewMuxEntry(handle_hello, "/hello"),
		web.NewMuxEntry(handle_helloworldagain, "/helloworldagain"),
		web.NewMuxEntry(handle_helloworld, "/helloworld"),
	}
	newmuxer := web.NewLongestPrefixRouter(mux_entries, fallback_handler)

	run_test := func(path string, expected_output string) {

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}
		newmuxer.ServeHTTP(rr, req)

		// Check the response body is what we expect.
		expected := expected_output
		if rr.Body.String() != expected {
			fmt.Println(len(rr.Body.String()), len(expected))
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}

	run_test("/hel.lo/", "MyCustomMuxer says: Invalid URL path.\n")

	run_test("/hello", "hello!")
	run_test("/hello/", "hello!")
	run_test("/helloa", "hello!")
	run_test("/helloworl", "hello!")

	run_test("/helloworld", "hello world!")
	run_test("/helloworlda", "hello world!")
	run_test("/helloworldagai", "hello world!")

	run_test("/helloworldagain", "hello world again!")
	run_test("/helloworldagaina", "hello world again!")

	run_test("/hell", "fallback!")
}
