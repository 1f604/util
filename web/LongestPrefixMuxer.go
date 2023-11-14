// I wrote this custom muxer because I wanted complete control over what happens - I wanted something simpler than http.ServeMux, which I think does too much

package util

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	web_types "github.com/1f604/util/web_types"
)

type MuxEntry struct {
	handler http.Handler
	prefix  string
}

// In Go, it is valid to call a method on a nil pointer!!!
func (p *MuxEntry) Length() int {
	if p == nil {
		return 0
	} else {
		return len(p.prefix)
	}
}

type LongestPrefixRouter struct {
	mux_entries     []*MuxEntry // unsored slice of entries
	fallbackHandler http.Handler
}

func NewMuxEntry(handler http.HandlerFunc, prefix string) *MuxEntry {
	return &MuxEntry{prefix: prefix, handler: handler}
}

func NewLongestPrefixRouter(entries []*MuxEntry, fallback_handler http.HandlerFunc) *LongestPrefixRouter {
	// Sanity check
	if len(entries) == 0 {
		panic("Empty list provided to NewLongestPrefixRouter")
	}

	return &LongestPrefixRouter{mux_entries: entries, fallbackHandler: fallback_handler}
}

func IsValidURL(url_path string) bool { // Accept paths like "/", "/root/", and "/root"
	_, err := web_types.PosixValidatedFullURLPath(url_path)
	if err == nil { // accept it
		return true
	}
	_, err = web_types.PosixValidatedURLDirPath(url_path)
	if err == nil { //nolint: gosimple // Keep it like this because we might want to add more rules later.
		return true
	}
	return false
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the RequestURI.
func (mux *LongestPrefixRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Reject invalid URL paths
	if !IsValidURL(r.URL.Path) {
		log.Printf("URL path %s is invalid.", r.URL.Path)
		http.Error(w, "MyCustomMuxer says: Invalid URL path.", http.StatusBadRequest)
		return
	}
	h := mux.match(r.URL.Path) // Will return default handler if no match found
	h.ServeHTTP(w, r)
}

// Find a handler given a path string.
// Most-specific (longest) pattern wins.
// If no match, returns fallback handler.
func (mux *LongestPrefixRouter) match(path string) http.Handler {
	fmt.Println("path:", path)
	var best_match *MuxEntry = nil
	// Check for longest valid match.
	for _, mux_entry := range mux.mux_entries {
		if strings.HasPrefix(path, mux_entry.prefix) && best_match.Length() < mux_entry.Length() {
			fmt.Println("path", path, "mux_entry.prefix", mux_entry.prefix)
			best_match = mux_entry
		}
	}
	if best_match != nil {
		fmt.Println("Best match:", best_match.prefix)
		return best_match.handler
	} else {
		fmt.Println("Best match is nil.")
	}
	return mux.fallbackHandler
}
