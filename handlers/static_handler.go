package handlers

import (
	"net/http"
	"os"
	"path"
	"strings"
)

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	staticPath := "./static"

	filePath := r.URL.Path

	if filePath == "/" {
		filePath = "/index.html"
	}

	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	fullPath := path.Join(staticPath, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.ServeFile(w, r, path.Join(staticPath, "index.html"))
		return
	}
	switch path.Ext(fullPath) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	http.ServeFile(w, r, fullPath)
}
