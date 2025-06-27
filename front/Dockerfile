# Используем nginx для раздачи статических файлов
FROM nginx:alpine

# Копируем статические файлы в директорию nginx
COPY . /usr/share/nginx/html/

# Копируем кастомную конфигурацию nginx (опционально)
# COPY nginx.conf /etc/nginx/nginx.conf

# Открываем порт 80
EXPOSE 80

# Nginx запускается автоматически 