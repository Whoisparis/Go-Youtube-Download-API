console.log('✅ script.js loaded successfully');

class YouTubeDownloader {
    constructor() {
        console.log('YouTubeDownloader constructor called');
        this.videoInfo = null;
        this.init();
    }

    init() {
        console.log('Initializing YouTubeDownloader...');
        this.bindEvents();
    }

    bindEvents() {
        console.log('Binding events...');

        const fetchBtn = document.getElementById('fetchInfoBtn');
        const urlInput = document.getElementById('videoUrl');

        if (fetchBtn && urlInput) {
            fetchBtn.addEventListener('click', (e) => {
                console.log('Fetch button clicked');
                this.fetchVideoInfo();
            });

            urlInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    console.log('Enter key pressed in URL input');
                    this.fetchVideoInfo();
                }
            });

            console.log('Events bound successfully');
        } else {
            console.error('Required elements not found for event binding');
        }
    }

    async fetchVideoInfo() {
        console.log('fetchVideoInfo called');
        const url = document.getElementById('videoUrl').value.trim();
        console.log('URL input value:', url);

        if (!url) {
            this.showError('Пожалуйста, введите URL видео');
            return;
        }

        if (!this.isValidYouTubeUrl(url)) {
            this.showError('Пожалуйста, введите корректный YouTube URL');
            return;
        }

        this.showLoading();
        this.hideError();

        try {
            console.log('Making API request to /api/video/info');
            const response = await fetch('/api/video/info', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ url })
            });

            console.log('API response status:', response.status);

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Ошибка при получении информации');
            }

            this.videoInfo = await response.json();
            console.log('Video info received:', this.videoInfo);
            this.displayVideoInfo();

        } catch (error) {
            console.error('Error in fetchVideoInfo:', error);
            this.showError(error.message);
        } finally {
            this.hideLoading();
        }
    }

    isValidYouTubeUrl(url) {
        const youtubeRegex = /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.?be)\/.+$/;
        return youtubeRegex.test(url);
    }

    displayVideoInfo() {
        console.log('Displaying video info');
        if (!this.videoInfo) return;

        try {
            // Заполняем информацию о видео
            document.getElementById('videoTitle').textContent = this.videoInfo.title;
            document.getElementById('videoAuthor').textContent = this.videoInfo.author;


            document.getElementById('videoDuration').textContent = this.formatDuration(this.videoInfo.duration);

            // Устанавливаем thumbnail
            const thumbnails = this.videoInfo.thumbnails || [];
            if (thumbnails.length > 0) {
                const thumbnailImg = document.getElementById('videoThumbnail');
                if (thumbnailImg) {
                    thumbnailImg.src = thumbnails[0].url;
                    thumbnailImg.onerror = () => {
                        console.warn('Thumbnail failed to load');
                    };
                }
            }

            // Создаем кнопки качества
            this.createQualityButtons();

            // Показываем блок с информацией
            this.showVideoInfo();

            console.log('Video info displayed successfully');

        } catch (error) {
            console.error('Error displaying video info:', error);
            this.showError('Ошибка при отображении информации о видео');
        }
    }

    createQualityButtons() {
        console.log('Creating quality buttons');
        const container = document.getElementById('qualityButtons');
        if (!container) {
            console.error('Quality buttons container not found');
            return;
        }

        container.innerHTML = '';

        if (!this.videoInfo.formats || this.videoInfo.formats.length === 0) {
            container.innerHTML = '<p>Нет доступных форматов для скачивания</p>';
            return;
        }

        this.videoInfo.formats.forEach(format => {
            const button = document.createElement('button');
            button.className = 'quality-btn';

            if (format.mimeType && format.mimeType.includes('audio')) {
                button.classList.add('audio');
            }

            const qualityLabel = format.qualityLabel || format.quality || 'Unknown';
            const fileSize = format.contentLength ? this.formatFileSize(format.contentLength) : 'N/A';

            button.innerHTML = `
                <div>${qualityLabel}</div>
                <small>${fileSize}</small>
            `;

            // ВАЖНО: Передаем itag как строку (String)
            button.addEventListener('click', () => {
                console.log('Quality button clicked, itag:', format.itag, 'type:', typeof format.itag);
                this.downloadVideo(format.itag.toString()); // Преобразуем в строку
            });

            container.appendChild(button);
        });

        console.log('Quality buttons created:', this.videoInfo.formats.length);
    }

    async downloadVideo(itag) {
        console.log('Downloading video with itag:', itag, 'type:', typeof itag);
        if (!this.videoInfo) return;

        this.showDownloadProgress();

        try {
            const url = document.getElementById('videoUrl').value.trim();

            // ВАЖНО: Отправляем itag как строку
            const requestData = {
                url: url,
                itag: itag.toString() // Убеждаемся, что это строка
            };

            console.log('Sending download request:', requestData);

            const response = await fetch('/api/video/download', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestData)
            });

            console.log('Download response status:', response.status);

            if (!response.ok) {
                let errorMessage = 'Ошибка при скачивании видео';
                try {
                    const errorData = await response.json();
                    errorMessage = errorData.message || errorMessage;
                    console.log('Error details:', errorData);
                } catch (e) {
                    errorMessage = `HTTP ${response.status}: ${response.statusText}`;
                }
                throw new Error(errorMessage);
            }

            const blob = await response.blob();
            console.log('Blob received, size:', blob.size);

            if (blob.size === 0) {
                throw new Error('Получен пустой файл');
            }

            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;

            // Находим формат для определения расширения
            const format = this.videoInfo.formats.find(f => f.itag.toString() === itag.toString());
            const extension = format?.mimeType?.includes('mp4') ? '.mp4' :
                format?.mimeType?.includes('webm') ? '.webm' :
                    format?.mimeType?.includes('m4a') ? '.m4a' : '.mp4';

            a.download = `${this.videoInfo.title}${extension}`;

            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);

            window.URL.revokeObjectURL(downloadUrl);

            this.hideDownloadProgress();
            console.log('Download completed successfully');

        } catch (error) {
            console.error('Download error:', error);
            this.hideDownloadProgress();
            this.showError(error.message);
        }
    }

    formatDuration(duration) {
        try {
            const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?/);
            if (!match) return duration;

            const hours = parseInt(match[1] || 0);
            const minutes = parseInt(match[2] || 0);
            const seconds = parseInt(match[3] || 0);

            if (hours > 0) {
                return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            } else {
                return `${minutes}:${seconds.toString().padStart(2, '0')}`;
            }
        } catch (error) {
            return duration;
        }
    }

    formatFileSize(bytes) {
        if (!bytes) return 'N/A';

        const sizes = ['Б', 'КБ', 'МБ', 'ГБ'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    // Вспомогательные методы для показа/скрытия элементов
    showLoading() {
        const loading = document.getElementById('loading');
        if (loading) loading.classList.remove('hidden');
        const btn = document.getElementById('fetchInfoBtn');
        if (btn) btn.disabled = true;
        console.log('Loading shown');
    }

    hideLoading() {
        const loading = document.getElementById('loading');
        if (loading) loading.classList.add('hidden');
        const btn = document.getElementById('fetchInfoBtn');
        if (btn) btn.disabled = false;
        console.log('Loading hidden');
    }

    showError(message) {
        const errorElement = document.getElementById('error');
        const errorMessage = document.getElementById('errorMessage');
        if (errorElement && errorMessage) {
            errorElement.classList.remove('hidden');
            errorMessage.textContent = message;
        }
        console.error('Error shown:', message);
    }

    hideError() {
        const errorElement = document.getElementById('error');
        if (errorElement) errorElement.classList.add('hidden');
        console.log('Error hidden');
    }

    showVideoInfo() {
        const videoInfo = document.getElementById('videoInfo');
        if (videoInfo) videoInfo.classList.remove('hidden');
        console.log('Video info shown');
    }

    hideVideoInfo() {
        const videoInfo = document.getElementById('videoInfo');
        if (videoInfo) videoInfo.classList.add('hidden');
        console.log('Video info hidden');
    }

    showDownloadProgress() {
        const progress = document.getElementById('downloadProgress');
        if (progress) progress.classList.remove('hidden');
        this.updateProgress(0);
        console.log('Download progress shown');
    }

    hideDownloadProgress() {
        const progress = document.getElementById('downloadProgress');
        if (progress) progress.classList.add('hidden');
        console.log('Download progress hidden');
    }

    updateProgress(percent) {
        const progressFill = document.querySelector('.progress-fill');
        const progressPercent = document.getElementById('progressPercent');

        if (progressFill) progressFill.style.width = percent + '%';
        if (progressPercent) progressPercent.textContent = percent + '%';

        console.log('Progress updated:', percent + '%');
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM fully loaded and parsed');
    try {
        new YouTubeDownloader();
        console.log('YouTubeDownloader initialized successfully');
    } catch (error) {
        console.error('Failed to initialize YouTubeDownloader:', error);
    }
});

// Также инициализируем при полной загрузке страницы
window.addEventListener('load', () => {
    console.log('Window fully loaded');
});

// Обработчик ошибок
window.addEventListener('error', (event) => {
    console.error('Global error:', event.error);
});