package main

import (
	"Youtube-download-API/handlers"
	"Youtube-download-API/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	"os"
	"path/filepath"
	"strings"
)

func main() {

	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORSMiddleware)

	// Создаем папку для статических файлов если не существует
	if err := os.MkdirAll("./static", 0755); err != nil {
		log.Printf("Warning: Cannot create static directory: %v", err)
	}

	// API routes
	youtubeHandler := handlers.NewYoutubeHandler()
	youtubeHandler.RegisterRoutes(router)

	// Статические файлы с правильными MIME types
	router.PathPrefix("/static/").HandlerFunc(staticFileHandler)

	// Главная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})

	log.Println("Starting server on :8080")
	log.Println("API endpoints:")
	log.Println("GET /api/health")
	log.Println("POST /api/video/info")
	log.Println("POST /api/video/download")
	log.Println("GET /download?url=YOUTUBE_URL")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err.Error())
	}
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем путь к файлу
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	if filePath == "" {
		http.NotFound(w, r)
		return
	}

	fullPath := filepath.Join("./static", filePath)

	// Проверяем существует ли файл
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Устанавливаем правильные MIME types
	switch filepath.Ext(fullPath) {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
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
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	http.ServeFile(w, r, fullPath)

)

func main() {

	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORSMiddleware)

	youtubeHandler := handlers.NewYoutubeHandler()
	youtubeHandler.RegisterRoutes(router)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Println("Starting server on :8080")
	log.Println("API endpoints:")
	log.Println("GET /api/health")
	log.Println("POST /api/video/info")
	log.Println("POST /api/video/download")
	log.Println("GET /download?url=YOUTUBE_URL")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err.Error())
	}

}
