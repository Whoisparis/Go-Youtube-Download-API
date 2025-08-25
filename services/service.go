package services

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type YoutubeService struct {
	client *youtube.Client
}

func NewYoutubeService() *YoutubeService {

	httpclient := &http.Client{
		Timeout: time.Minute * 10,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	return &YoutubeService{
		client: &youtube.Client{
			HTTPClient: httpclient,
		},
	}
}

func (s *YoutubeService) GetVideoInfo(ctx context.Context, url string) (*youtube.Video, error) {
	video, err := s.client.GetVideoContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %v", err)
	}
	return video, nil
}

func (s *YoutubeService) DownloadVideo(ctx context.Context, video *youtube.Video, itagStr string, quality string) (*youtube.Format, io.ReadCloser, error) {
	var format *youtube.Format


	log.Printf("Looking for format: itag=%s, quality=%s", itagStr, quality)

	// Приоритет 1: Поиск по ITAG (если указан)
	if itagStr != "" {
		// Конвертируем строку в число

	if itagStr != "" {
		itag, err := strconv.Atoi(itagStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid itag format: must be a number, got '%s'", itagStr)
		}


		log.Printf("Searching for itag: %d", itag)
		for i := range video.Formats {
			if video.Formats[i].ItagNo == itag {
				format = &video.Formats[i]
				log.Printf("Found format by itag: %+v", format)
		for i := range video.Formats {
			if video.Formats[i].ItagNo == itag {
				format = &video.Formats[i]
				break
			}
		}
		if format == nil {
			return nil, nil, fmt.Errorf("format with itag %d not found. Available itags: %v",
				itag, getAvailableItags(video.Formats))
		}
	}

	// Приоритет 2: Поиск по качеству (если указано)
	if format == nil && quality != "" {
		log.Printf("Searching for quality: %s", quality)
		formats := video.Formats.Quality(quality)
		if len(formats) > 0 {
			format = &formats[0]
			log.Printf("Found format by quality: %+v", format)
		} else {
			return nil, nil, fmt.Errorf("format with quality '%s' not found. Available qualities: %v",
				quality, getAvailableQualities(video.Formats))
		}
	}

	// Приоритет 3: Берем лучший доступный формат с аудио
	if format == nil {
		log.Printf("Selecting best available format with audio")
			return nil, nil, fmt.Errorf("format with itag %s not found", itagStr)
		}
	}

	if format == nil && quality == "" {
		formats := video.Formats.Quality(quality)
		if len(formats) > 0 {
			format = &formats[0]
		} else {
			return nil, nil, fmt.Errorf("no formats found for quality '%s'", quality)
		}
	}

	if format == nil {
		formats := video.Formats.WithAudioChannels()
		if len(formats) == 0 {
			return nil, nil, fmt.Errorf("no suitable format with audio found")
		}
		format = &formats[0]
		log.Printf("Selected best format: %+v", format)
	}

	// Получаем поток видео
	stream, _, err := s.client.GetStreamContext(ctx, video, format)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get video stream: %w", err)
	}

	stream, _, err := s.client.GetStreamContext(ctx, video, format)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get video stream: %v", err)
	}

	return format, stream, nil
}

func (s *YoutubeService) SanitizeFilename(filename string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	if len(filename) > 100 {
		filename = filename[:100]
	}

	return filename
}

func (s *YoutubeService) GetFileExtension(mimeType string) string {
	switch {
	case strings.Contains(mimeType, ".mp4"):
		return ".mp4"
	case strings.Contains(mimeType, ".webm"):
		return ".webm"
	case strings.Contains(mimeType, ".m4a"):
		return ".m4a"
	case strings.Contains(mimeType, ".mp3"):
		return ".mp3"
	default:
		return ".mp4"
	}
}

func getAvailableQualities(formats []youtube.Format) []string {
	qualities := make(map[string]bool)
	for _, f := range formats {
		if f.Quality != "" {
			qualities[f.Quality] = true
		}
		if f.QualityLabel != "" {
			qualities[f.QualityLabel] = true
		return ".video"
	}
}

func (s *YoutubeService) GetAvailableQualities(formats []youtube.Format) []string {
	qualities := make(map[string]bool, len(formats))
	for _, format := range formats {
		if format.QualityLabel != "" {
			qualities[format.QualityLabel] = true
		}
	}

	result := make([]string, 0, len(qualities))
	for q := range qualities {
		result = append(result, q)
	}
	return result
}

func getAvailableItags(formats []youtube.Format) []int {
	itags := make([]int, 0)
	for _, f := range formats {
		itags = append(itags, f.ItagNo)
	}
	return itags
}
	for quality := range qualities {
		result = append(result, quality)
	}
	return result
}