package api

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func serveWeb(w http.ResponseWriter, r *http.Request, root string) {
	if root == "" || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
		http.NotFound(w, r)
		return
	}
	requestPath, err := url.PathUnescape(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	requestPath = strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if requestPath == "." || requestPath == "" {
		requestPath = "index.html"
	}
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	target := filepath.Join(cleanRoot, filepath.FromSlash(requestPath))
	if !strings.HasPrefix(target, cleanRoot+string(filepath.Separator)) || strings.Contains(requestPath, "..") {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		if filepath.Ext(requestPath) != "" {
			http.NotFound(w, r)
			return
		}
		requestPath = "index.html"
		target = filepath.Join(cleanRoot, filepath.FromSlash(requestPath))
		info, err = os.Stat(target)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
	}
	file, err := os.Open(target)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	http.ServeContent(w, r, filepath.Base(target), info.ModTime(), file)
}
