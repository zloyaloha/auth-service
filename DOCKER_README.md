# Docker инструкции для Ticket App

Этот документ содержит инструкции по запуску приложения Ticket App с использованием Docker.

## Требования

- Docker
- Docker Compose

## Структура сервисов

Приложение состоит из следующих сервисов:

- **postgres** - База данных PostgreSQL
- **auth-service** - gRPC сервис аутентификации (порт 50051)
- **gateway** - HTTP Gateway сервис (порт 8080)
- **frontend** - Веб-интерфейс (порт 80)

## Быстрый старт

### 1. Сборка и запуск всех сервисов

```bash
# Сборка образов и запуск
make up-build

# Или пошагово:
make build
make up
```

### 2. Проверка статуса

```bash
make status
```

### 3. Просмотр логов

```bash
# Все сервисы
make logs

# Конкретный сервис
make logs-auth
make logs-gateway
make logs-frontend
make logs-postgres
```

## Доступ к сервисам

После запуска сервисы будут доступны по следующим адресам:

- **Frontend**: http://localhost
- **Gateway API**: http://localhost:8080
- **Auth Service (gRPC)**: localhost:50051
- **PostgreSQL**: localhost:5432

## Управление контейнерами

### Основные команды

```bash
# Остановка всех сервисов
make down

# Перезапуск
make restart

# Остановка с удалением volumes
make down-volumes
```

### Очистка

```bash
# Очистка неиспользуемых образов и контейнеров
make clean

# Полная очистка (включая volumes)
make clean-all
```

## Работа с базой данных

### Подключение к базе данных

```bash
make db-connect
```

### Резервное копирование

```bash
# Создание резервной копии
make db-backup

# Восстановление из резервной копии
make db-restore BACKUP_FILE=backup_20241201_120000.sql
```

### Выполнение миграций

```bash
make migrate
```

## Разработка

### Пересборка конкретного сервиса

```bash
# Пересборка auth-service
docker-compose build auth-service
docker-compose up -d auth-service

# Пересборка gateway
docker-compose build gateway
docker-compose up -d gateway
```

### Просмотр логов в реальном времени

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f auth-service
```

## Конфигурация

### Переменные окружения

Основные переменные окружения находятся в файле `auth-service/configs/postgres.env`:

```
POSTGRES_PORT=5432
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=postgres
```

### Изменение портов

Для изменения портов отредактируйте секцию `ports` в `docker-compose.yaml`:

```yaml
ports:
  - "8080:8080"  # host_port:container_port
```

## Устранение неполадок

### Проверка состояния сервисов

```bash
docker-compose ps
```

### Проверка логов

```bash
docker-compose logs [service_name]
```

### Перезапуск с пересборкой

```bash
docker-compose down
docker-compose up -d --build
```

### Очистка и пересборка

```bash
make clean
make up-build
```

## Производительность

### Оптимизация образов

- Используется многоэтапная сборка для Go сервисов
- Минимальные базовые образы (alpine)
- Кэширование слоев Docker

### Мониторинг ресурсов

```bash
# Просмотр использования ресурсов
docker stats
```

## Безопасность

- Сервисы запускаются от непривилегированного пользователя
- Используется изолированная сеть Docker
- Данные PostgreSQL сохраняются в именованных volumes 