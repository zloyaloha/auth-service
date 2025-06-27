# Многоэтапная сборка для Go приложения
FROM golang:1.24.4-alpine AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/gateaway

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /app/main .
# Копируем конфигурационные файлы
COPY --from=builder /app/config ./config

# Меняем владельца файлов
RUN chown -R appuser:appgroup /root/

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт
EXPOSE 8080

# Запускаем приложение
# В Dockerfile для указания команды используется либо массив (exec-форма), либо строка (shell-форма).
# В вашем случае вы используете exec-форму, но передаёте весь аргумент как одну строку.
# Docker будет искать исполняемый файл с именем "./main --config=./config/config.yaml", а не "./main" с аргументом "--config=./config/config.yaml".
# Нужно разделить команду и аргументы в массиве:

CMD ["./main", "--config=./config/config.yaml"]