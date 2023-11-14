// I wrote this custom muxer because I wanted complete control over what happens - I wanted something simpler than http.ServeMux, which I think does too much

package util

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/1f604/util"
	web_types "github.com/1f604/util/web_types"
)

type HandlerEnum int

type HandlerType struct {
	type_name string
}

var EXACT_MATCH_HANDLER HandlerType = HandlerType{"exact match handler"}
var LONGEST_PREFIX_HANDLER HandlerType = HandlerType{"longest prefix handler"}

type MuxEntry struct {
	handler      http.Handler
	prefix       string
	handler_type util.HandlerTypeEnum
}

type HandlerPair struct {
	handler http.Handler
	prefix  string
}

// In Go, it is valid to call a method on a nil pointer!!!
func (p *HandlerPair) Length() int {
	if p == nil {
		return 0
	} else {
		return len(p.prefix)
	}
}

type LongestPrefixRouter struct {
	prefix_handlers      []*HandlerPair // unsored slice of entries
	exact_match_handlers map[string]http.Handler
	fallbackHandler      http.Handler
}

func NewMuxEntry(handler http.HandlerFunc, prefix string, handler_type util.HandlerTypeEnum) *MuxEntry {
	return &MuxEntry{prefix: prefix, handler: handler, handler_type: handler_type}
}

func NewLongestPrefixRouter(entries []*MuxEntry, fallback_handler http.HandlerFunc) *LongestPrefixRouter {
	// Sanity check
	if len(entries) == 0 {
		panic("Empty list provided to NewLongestPrefixRouter")
	}

	prefix_handlers := make([]*HandlerPair, 0)
	exact_match_handlers := make(map[string]http.Handler)
	for i := range entries {
		switch entries[i].handler_type.(type) {
		case util.LONGEST_PREFIX_HANDLER_t:
			prefix_handlers = append(prefix_handlers, &HandlerPair{handler: entries[i].handler, prefix: entries[i].prefix})
		case util.EXACT_MATCH_HANDLER_t:
			exact_match_handlers[entries[i].prefix] = entries[i].handler
		default:
			panic("Unrecognized handler_type in mux entry.")
		}
	}

	return &LongestPrefixRouter{
		prefix_handlers:      prefix_handlers,
		exact_match_handlers: exact_match_handlers,
		fallbackHandler:      fallback_handler,
	}
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
	// log request
	Nginx_Log_Received_Request(r)

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
	// First try to find an exact match
	result, ok := mux.exact_match_handlers[path]
	if ok {
		fmt.Println("exact matched!")
		return result
	}

	// If no exact match, then check for longest valid match.
	var best_match *HandlerPair = nil
	for _, mux_entry := range mux.prefix_handlers {
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
