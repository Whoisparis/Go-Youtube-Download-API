package models

import "github.com/kkdai/youtube/v2"

type GetVideoInfoRequest struct {
	URL string `json:"url"`
}
type DownloadRequest struct {
	URL     string `json:"url"`
	Itag    string `json:"itag"`
	Quality string `json:"quality"`
}

type VideoInfoResponse struct {
	Title       string       `json:"title"`
	Duration    string       `json:"duration"`
	Author      string       `json:"author"`
	Description string       `json:"description"`
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

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func ConvertToFormatInfo(format youtube.Format) FormatInfo {
	return FormatInfo{
		Itag:          format.ItagNo,
		Quality:       format.Quality,
		MimeType:      format.MimeType,
		AudioQuality:  format.AudioQuality,
		ContentLength: format.ContentLength,
		QualityLabel:  format.QualityLabel,
	}
}
