package handlers

import (
	"Youtube-download-API/models"
	"Youtube-download-API/services"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"time"
)

type YoutubeHandler struct {
	youtubeService *services.YoutubeService
	startTime      time.Time
}

func NewYoutubeHandler() *YoutubeHandler {
	return &YoutubeHandler{
		youtubeService: services.NewYoutubeService(),
		startTime:      time.Now(),
	}
}

func (h *YoutubeHandler) GetVideoInfoHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GetVideoInfoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.sendError(w, "URL is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	video, err := h.youtubeService.GetVideoInfo(ctx, req.URL)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formats := make([]models.FormatInfo, 0, len(video.Formats))
	for _, format := range video.Formats {
		if format.MimeType != "" && format.Quality != "" {
			formats = append(formats, models.ConvertToFormatInfo(format))
		}
	}

	response := models.VideoInfoResponse{
		Title:       video.Title,
		Duration:    video.Duration.String(),
		Author:      video.Author,
		Description: video.Description,
		Formats:     formats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		err.Error()
		return
	}
}

func (h *YoutubeHandler) DownloadVideoHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DownloadRequest

	if r.Method == http.MethodGet {
		req.URL = r.URL.Query().Get("url")
		req.Itag = r.URL.Query().Get("itag")
		req.Quality = r.URL.Query().Get("quality")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	if req.URL == "" {
		h.sendError(w, "URL is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	video, err := h.youtubeService.GetVideoInfo(ctx, req.URL)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	format, stream, err := h.youtubeService.DownloadVideo(ctx, video, req.Itag, req.Quality)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func(stream io.ReadCloser) {
		err := stream.Close()
		if err != nil {
			err.Error()
		}
	}(stream)

	filename := h.youtubeService.SanitizeFilename(video.Title) + h.youtubeService.GetFileExtension(format.MimeType)

	w.Header().Set("Content-Disposition", "attachment; filename="+filename+"\"")
	w.Header().Set("Content-Type", format.MimeType)
	if format.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(format.ContentLength, 10))
	}
	if _, err := io.Copy(w, stream); err != nil {
		h.sendError(w, "Failed to stream video"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *YoutubeHandler) HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "Ok",
		"version": "1.0.0",
	})
	if err != nil {
		h.sendError(w, "Failed to write health check", http.StatusInternalServerError)
	}
}

func (h *YoutubeHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.sendError(w, "Failed to write health check", http.StatusInternalServerError)
		return
	}

}

func (h *YoutubeHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/healt", h.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/api/video/info", h.GetVideoInfoHandler).Methods("POST")
	router.HandleFunc("/api/video/download", h.DownloadVideoHandler).Methods("GET", "POST")
	router.HandleFunc("/download", h.DownloadVideoHandler).Methods("GET", "POST")

	router.HandleFunc("/api/video/info", h.optionsHandler).Methods("OPTIONS")
	router.HandleFunc("/api/video/download", h.optionsHandler).Methods("OPTIONS")
	router.HandleFunc("/api/video/download", h.optionsHandler).Methods("OPTIONS")
}

func (h *YoutubeHandler) optionsHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
