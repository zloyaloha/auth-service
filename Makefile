.PHONY: build up down logs clean restart status

# Сборка всех образов
build:
	docker-compose build

# Запуск всех сервисов
up:
	docker-compose up -d

# Запуск с пересборкой образов
up-build:
	docker-compose up -d --build

# Остановка всех сервисов
down:
	docker-compose down

# Остановка с удалением volumes
down-volumes:
	docker-compose down -v

# Просмотр логов всех сервисов
logs:
	docker-compose logs -f

# Просмотр логов конкретного сервиса
logs-auth:
	docker-compose logs -f auth-service

logs-gateway:
	docker-compose logs -f gateway

logs-frontend:
	docker-compose logs -f frontend

logs-postgres:
	docker-compose logs -f postgres

# Перезапуск всех сервисов
restart:
	docker-compose restart

# Статус сервисов
status:
	docker-compose ps

# Очистка неиспользуемых образов и контейнеров
clean:
	docker system prune -f
	docker volume prune -f

# Полная очистка (включая volumes)
clean-all:
	docker system prune -a -f
	docker volume prune -f

# Выполнение миграций
migrate:
	docker-compose exec auth-service ./main migrate

# Подключение к базе данных
db-connect:
	docker-compose exec postgres psql -U admin -d postgres

# Создание резервной копии базы данных
db-backup:
	docker-compose exec postgres pg_dump -U admin postgres > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Восстановление базы данных из резервной копии
db-restore:
	docker-compose exec -T postgres psql -U admin -d postgres < $(BACKUP_FILE) 