package util

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/1f604/util"
	web_types "github.com/1f604/util/web_types"
)

// The user supplies a map of URL prefix to file system directory paths to the server and the server will map each URL, replacing the prefix with the file system directory path from the map.
// As a safety precaution, the SafelyServeFile function will refuse to serve any files which does not have the xattr set. It will also refuse to list directories.
//
// I realize that the amount of logging in this function is not going to serve everyone's needs
// So if you want more or less logging (or you want something else), then copy and modify it so that it does what you want.
// This function has been manually tested to verify that it handles all edge cases correctly.
// As a safety precaution, this function will refuse to serve files whose permissions are not set to 777.
func SafelyServeFile(w http.ResponseWriter, r *http.Request, url_to_dir_map web_types.URLPrefixToFileSystemDirectoryMap) { //nolint:funlen // it's fine
	Nginx_Log_Received_Request(r)

	if r.Method != http.MethodGet { // we only support HTTP GET requests
		log.Print("Method not supported.")
		http.Error(w, "Method not supported.", http.StatusBadRequest)
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
