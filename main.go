package main

import (
	"Youtube-download-API/handlers"
	"Youtube-download-API/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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
