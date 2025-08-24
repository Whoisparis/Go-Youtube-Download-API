package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"net/http"
	"strings"
)

type VideoInfo struct {
	Title       string       `json:"title"`
	Duration    string       `json:"duration"`
	Author      string       `json:"author"`
	Description string       `json:"Description"`
	Formats     []FormatInfo `json:"formats"`
}

type FormatInfo struct {
	Itag          int    `json:"itag"`
	Quality       string `json:"quality"`
	MimeType      string `json:"mimetype"`
	AudioQuality  string `json:"audio_quality"`
	ContentLength int64  `json:"content_length"`
	QualityLabel  string `json:"quality_label"`
}

type DownloadHandler struct {
	client *youtube.Client
}

func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{
		client: &youtube.Client{},
	}
}

func (h *DownloadHandler) GetVideoInfo(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	video, err := h.client.GetVideo(request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formats := make([]FormatInfo, 0)

	for _, format := range video.Formats {
		formats = append(formats, FormatInfo{
			Itag:          format.ItagNo,
			Quality:       format.Quality,
			MimeType:      format.MimeType,
			AudioQuality:  format.AudioQuality,
			ContentLength: format.ContentLength,
			QualityLabel:  format.QualityLabel,
		})
	}

	response := VideoInfo{
		Title:       video.Title,
		Duration:    video.Duration.String(),
		Author:      video.Author,
		Description: video.Description,
		Formats:     formats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *DownloadHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL     string `json:"url"`
		Quality string `json:"quality"`
		Itag    int    `json:"itag"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	video, err := h.client.GetVideo(request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var format *youtube.Format
	if request.Itag != 0 {
		for _, f := range video.Formats {
			if f.ItagNo == request.Itag {
				format = &f
				break
			}
		}
	}

	if format == nil {
		formats := video.Formats.WithAudioChannels()
		if len(formats) > 0 {
			format = &formats[0]
		} else {
			http.Error(w, "no formats found", http.StatusBadRequest)
			return
		}
	}

	stream, _, err := h.client.GetStream(video, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(stream io.ReadCloser) {
		err := stream.Close()
		if err != nil {
			log.Println(err)
		}
	}(stream)

	filename := sanitizeFilename(video.Title) + ".mp4"
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "video/mp4")

	_, err = io.Copy(w, stream)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func sanitizeFilename(filename string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	return filename
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	handler := NewDownloadHandler()

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	r.HandleFunc("/api/video/info", handler.GetVideoInfo).Methods("POST")
	r.HandleFunc("/api/video/download", handler.DownloadVideo).Methods("POST")

	r.PathPrefix("/downloads/").Handler(http.StripPrefix("/downloads/", http.FileServer(http.Dir("./downloads"))))

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
