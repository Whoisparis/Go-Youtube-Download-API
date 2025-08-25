package services

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
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

	if itagStr != "" {
		itag, err := strconv.Atoi(itagStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid itag format: must be a number, got '%s'", itagStr)
		}

		for i := range video.Formats {
			if video.Formats[i].ItagNo == itag {
				format = &video.Formats[i]
				break
			}
		}
		if format == nil {
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
	for quality := range qualities {
		result = append(result, quality)
	}
	return result
}
