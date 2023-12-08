package util

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/1f604/util"
	web_types "github.com/1f604/util/web_types"
)

// The user supplies a map of URL prefix to file system directory paths to the server and the server will map each URL, replacing the prefix with the file system directory path from the map.
// As a safety precaution, the SafelyServeFile function will refuse to serve any files which does not have the correct xattr set. It will also refuse to list directories.
//
// I realize that the amount of logging in this function is not going to serve everyone's needs
// So if you want more or less logging (or you want something else), then copy and modify it so that it does what you want.
// This function has been manually tested to verify that it handles all edge cases correctly.
func SafelyServeFile(w http.ResponseWriter, r *http.Request, url_to_dir_map web_types.URLPrefixToFileSystemDirectoryMap, log_request bool) { //nolint:funlen // it's fine
	if log_request {
		Nginx_Log_Received_Request("SafelyServeFile", r)
	}

	if r.Method != http.MethodGet { // we only support HTTP GET requests
		log.Print("Method not allowed.")
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
		return
	}

	urlpath_str := r.URL.Path
	log.Printf("Received request for %s", urlpath_str)

	defer func() {
		log.Print("Finished handling request.")
	}()

	// Validate URL path
	posix_validated_url_path, err := web_types.PosixValidatedFullURLPath(urlpath_str)
	if err != nil || posix_validated_url_path == nil {
		// return a BadRequest error saying the URL path was not valid
		log.Printf("URL path %s is invalid.", urlpath_str)
		http.Error(w, "Invalid URL path.", http.StatusBadRequest)
		return
	}

	// Last minute sanity check
	dirpath, _ := filepath.Split(r.URL.Path)
	if strings.Contains(dirpath, ".") {
		log.Printf("Panic: URL path %s contains a dot.", dirpath)
		panic("This should never happen.")
	}

	// we validated the URL path so now try to convert it to the file system path
	convertedfspath, err := web_types.Convert_URLPath_To_Full_FileSystem_Path(*posix_validated_url_path, url_to_dir_map)
	if err != nil || convertedfspath == nil {
		// return a NotFound error saying we do not serve the URL path
		log.Print("URL prefix not matched.")
		http.Error(w, "Requested URL prefix was not found in the list of URL prefixes in the server configuration.", http.StatusNotFound)
		return
	}

	// Last minute sanity check
	dirpath, _ = filepath.Split(*convertedfspath.FileSystemPath)
	if strings.Contains(dirpath, "..") {
		log.Printf("Panic: file system path %s contains a dot dot.", dirpath)
		panic("This should never happen.")
	}

	// We converted the URL path to a file system path, now try to serve it
	log.Printf("URL path %s maps to file system path %s", urlpath_str, *convertedfspath.FileSystemPath)
	log.Printf("Now opening file %s", *convertedfspath.FileSystemPath)
	f, err := os.Open(*convertedfspath.FileSystemPath)
	if err != nil {
		log.Print("Failed to open file:", err)
		http.Error(w, "Failed to open file.", http.StatusNotFound)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		log.Print("Failed to stat file:", err)
		http.Error(w, "Failed to stat file.", http.StatusInternalServerError)
		return
	}

	// Check file xattr to determine if we can serve it or not.
	out_buf := make([]byte, 300) //nolint:gomnd // this is technically not safe. what could go wrong?
	_, err = util.Getxattr(*convertedfspath.FileSystemPath, util.XATTR_1F604_FILESERVER_CAN_BE_SERVED, out_buf)
	if err != nil {
		log.Print("Refusing to serve file because the xattr is not set:", err)
		http.Error(w, "File does not have the required xattr set or xattr value is incorrect.", http.StatusInternalServerError)
		return
	}
	if !bytes.Equal(out_buf[:4], []byte("true")) {
		log.Print("Refusing to serve file because the xattr value is incorrect:", string(out_buf[:4]))
		http.Error(w, "File does not have the required xattr set or xattr value is incorrect.", http.StatusInternalServerError)
		return
	}

	// Check if it's a directory
	if d.IsDir() {
		log.Printf("%s is a directory.", *convertedfspath.FileSystemPath)
		fmt.Fprintf(w, "This is a directory.")
		return
	}

	// ServeContent will check modification time
	log.Printf("Now serving file %s", *convertedfspath.FileSystemPath)
	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

// Simple function for serving embedded files
// Example usage: If you're serving a directory called "static" as "resources", then you want to call it like this:
//
// SafelyServeFileEmbedded(w, r, "resources/", "static/", embedfs, false)
func SafelyServeFileEmbedded(w http.ResponseWriter, r *http.Request, url_prefix string, fs_prefix string, embedfs embed.FS, log_request bool) { //nolint:funlen // it's fine
	if log_request {
		Nginx_Log_Received_Request("SafelyServeFileStatic", r)
	}

	if r.Method != http.MethodGet { // we only support HTTP GET requests
		log.Print("Method not allowed.")
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
		return
	}

	urlpath_str := r.URL.Path
	log.Printf("Received request for %s", urlpath_str)

	defer func() {
		log.Print("Finished handling request.")
	}()

	// Check prefix
	if !strings.HasPrefix(urlpath_str, url_prefix) {
		// return a BadRequest error saying the URL prefix is wrong
		log.Printf("URL prefix is invalid for url path %s", urlpath_str)
		log.Printf("Expected URL prefix %s", url_prefix)
		http.Error(w, "URL prefix is invalid.", http.StatusInternalServerError)
		return
	}
	// Now remove the prefix from the string
	short_urlpath_str := urlpath_str[len(url_prefix):]

	// Now map it to fs path
	fs_path := fs_prefix + short_urlpath_str

	// Validate fs path
	posix_validated_fs_path, err := web_types.PosixValidatedFullURLPath(fs_path)
	if err != nil || posix_validated_fs_path == nil {
		// return a BadRequest error saying the URL path was not valid
		log.Printf("URL path %s is invalid. Error: %v", urlpath_str, err)
		http.Error(w, "Invalid URL path.", http.StatusBadRequest)
		return
	}

	// Last minute sanity check
	dirpath, _ := filepath.Split(posix_validated_fs_path.URLPath)
	if strings.Contains(dirpath, ".") {
		log.Printf("Panic: URL path %s contains a dot.", dirpath)
		panic("This should never happen.")
	}

	// Now open the file
	log.Printf("URL path %s maps to file system path %s", urlpath_str, posix_validated_fs_path.URLPath)
	log.Printf("Now opening file %s", posix_validated_fs_path.URLPath)
	f, err := embedfs.Open(posix_validated_fs_path.URLPath)
	if err != nil {
		log.Print("Failed to open file:", err)
		http.Error(w, "Failed to open file.", http.StatusNotFound)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		log.Print("Failed to stat file:", err)
		http.Error(w, "Failed to stat file.", http.StatusInternalServerError)
		return
	}

	// ServeContent will check modification time
	log.Printf("Now serving file %s", posix_validated_fs_path.URLPath)
	http.ServeContent(w, r, d.Name(), d.ModTime(), f.(io.ReadSeeker))
}
