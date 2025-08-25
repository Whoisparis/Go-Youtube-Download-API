# Build stage
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o youtube-downloader .

# Final stage
FROM alpine:3.18

# Устанавливаем зависимости времени выполнения
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    && update-ca-certificates

# Создаем пользователя для безопасности
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Создаем директории
RUN mkdir -p /app/static /app/downloads \
    && chown -R appuser:appgroup /app

# Переключаемся на пользователя appuser
USER appuser

# Рабочая директория
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder --chown=appuser:appgroup /app/youtube-downloader .
COPY --chown=appuser:appgroup static/ ./static/

# Создаем volume для загрузок
VOLUME /app/downloads

# Открываем порт
EXPOSE 8080

# Переменные окружения
ENV GO_ENV=production
ENV PORT=8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Запускаем приложение
CMD ["./youtube-downloader"]