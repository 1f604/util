package util_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1f604/util"
	web "github.com/1f604/util/web"
)

func handle_hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!")) //nolint: errcheck // no need to check error here
}

func handle_hello2(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("yoooo !!!")) //nolint: errcheck // no need to check error here
}

func handle_hello_exact(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello exact!")) //nolint: errcheck // no need to check error here
}

func handle_helloworld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!")) //nolint: errcheck // no need to check error here
}

func handle_helloworldagain(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world again!")) //nolint: errcheck // no need to check error here
}

func fallback_handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("fallback!")) //nolint: errcheck // no need to check error here
}

func Test_LongestPrefixMuxer(t *testing.T) {
	t.Parallel()

	mux_entries := []*web.MuxEntry{
		web.NewMuxEntry("example.com", handle_hello_exact, "/hello", util.EXACT_MATCH_HANDLER),
		web.NewMuxEntry("example.com", handle_hello, "/hello", util.LONGEST_PREFIX_HANDLER),
		web.NewMuxEntry("example.com", handle_helloworldagain, "/helloworldagain", util.LONGEST_PREFIX_HANDLER),
		web.NewMuxEntry("example.com", handle_helloworld, "/helloworld", util.LONGEST_PREFIX_HANDLER),
		web.NewMuxEntry("example2.com", handle_hello2, "/hello", util.EXACT_MATCH_HANDLER),
	}
	newmuxer := web.NewLongestPrefixRouter(mux_entries, fallback_handler, false)

	run_test := func(path string, expected_output string) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, path, nil) //nolint: noctx // dont need a context here.
		if err != nil {
			t.Fatal(err)
		}
		newmuxer.ServeHTTP(rr, req)

		// Check the response body is what we expect.
		expected := expected_output
		if rr.Body.String() != expected {
			// fmt.Println(len(rr.Body.String()), len(expected))
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}

	run_test("https://example.com/hel.lo/", "MyCustomMuxer says: Invalid URL path.\n")

	run_test("https://example.com/hello", "hello exact!")
	run_test("https://example2.com/hello", "yoooo !!!")
	run_test("https://example.com/hello/", "hello!")
	run_test("https://example.com/helloa", "hello!")
	run_test("https://example.com/helloworl", "hello!")

	run_test("https://example.com/helloworld", "hello world!")
	run_test("https://example.com/helloworlda", "hello world!")
	run_test("https://example.com/helloworldagai", "hello world!")

	run_test("https://example.com/helloworldagain", "hello world again!")
	run_test("https://example.com/helloworldagaina", "hello world again!")

	run_test("https://example.com/hell", "fallback!")
}
