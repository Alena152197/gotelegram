# Глава 10: Развёртывание бота

В этой главе мы научимся разворачивать бота в продакшене. Это включает в себя создание Docker-образов, настройку сервера и мониторинг.

---

## 1. Подготовка к продакшену

**Что нужно сделать:**
- Убрать debug-режим
- Настроить правильное логирование
- Настроить обработку ошибок
- Подготовить переменные окружения
- Настроить мониторинг

**Обновляем конфигурацию:**
```go
// В config.go добавляем окружение
type Config struct {
    Environment string `envconfig:"ENV" default:"development"` // development, production
    // ... остальные поля
}

// Проверяем окружение
if cfg.Environment == "production" {
    bot.Debug = false
    // Отключаем подробное логирование
}
```

---

## 2. Создаём Dockerfile

**Создаём файл `Dockerfile`:**
```dockerfile
# Этап сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Финальный этап
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из этапа сборки
COPY --from=builder /app/bot .

# Копируем миграции (если нужны)
COPY --from=builder /app/migrations ./migrations

# Запускаем приложение
CMD ["./bot"]
```

**Сборка образа:**
```bash
docker build -t telegram-bot .
```

**Запуск контейнера:**
```bash
docker run -d \
  --name telegram-bot \
  --env-file .env \
  telegram-bot
```

---

## 3. Docker Compose для полного стека

**Создаём `docker-compose.yml`:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: telegram-bot-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: telegram_bot
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    container_name: telegram-bot-redis
    ports:
      - "6379:6379"

  bot:
    build: .
    container_name: telegram-bot
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    volumes:
      - ./logs:/root/logs  # Монтируем папку для логов

volumes:
  postgres_data:
```

**Запуск:**
```bash
docker-compose up -d
```

**Просмотр логов:**
```bash
docker-compose logs -f bot
```

---

## 4. Systemd service (для Linux сервера)

**Создаём файл `/etc/systemd/system/telegram-bot.service`:**
```ini
[Unit]
Description=Telegram Bot
After=network.target postgresql.service

[Service]
Type=simple
User=telegram-bot
WorkingDirectory=/opt/telegram-bot
EnvironmentFile=/opt/telegram-bot/.env
ExecStart=/opt/telegram-bot/bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Управление сервисом:**
```bash
sudo systemctl start telegram-bot
sudo systemctl stop telegram-bot
sudo systemctl status telegram-bot
sudo systemctl enable telegram-bot  # Автозапуск при загрузке
```

---

## 5. Мониторинг и логирование

### Структурированные логи

**Используем zap для логирования в файл:**
```go
func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"stdout", cfg.File}
    
    return config.Build()
}
```

### Health check

**Создаём HTTP endpoint для проверки здоровья:**
```go
// Добавляем простой HTTP-сервер для health check
func startHealthCheckServer() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    go http.ListenAndServe(":8080", nil)
}
```

---

## 6. Обновление бота без простоя

**Стратегия:**
1. Собрать новый образ
2. Остановить старый контейнер
3. Запустить новый контейнер

**Скрипт обновления:**
```bash
#!/bin/bash
# update.sh

echo "Сборка образа..."
docker build -t telegram-bot:latest .

echo "Остановка старого контейнера..."
docker stop telegram-bot

echo "Удаление старого контейнера..."
docker rm telegram-bot

echo "Запуск нового контейнера..."
docker run -d \
  --name telegram-bot \
  --env-file .env \
  --restart unless-stopped \
  telegram-bot:latest

echo "Обновление завершено!"
```

---

## 7. Backup базы данных

**Создаём скрипт бэкапа:**
```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/telegram_bot_$DATE.sql"

docker exec telegram-bot-postgres pg_dump -U postgres telegram_bot > $BACKUP_FILE

# Сжимаем бэкап
gzip $BACKUP_FILE

# Удаляем старые бэкапы (старше 7 дней)
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

---

## 8. CI/CD с GitHub Actions

**Создаём `.github/workflows/deploy.yml`:**
```yaml
name: Deploy Bot

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Build Docker image
      run: docker build -t telegram-bot .
    
    - name: Deploy to server
      uses: appleboy/ssh-action@master
      with:
        host: ${{ secrets.SERVER_HOST }}
        username: ${{ secrets.SERVER_USER }}
        key: ${{ secrets.SSH_KEY }}
        script: |
          cd /opt/telegram-bot
          docker-compose pull
          docker-compose up -d
```

---

## Типичные проблемы

**Проблема 1: Бот падает после перезапуска**

**Решение:** Используйте `restart: unless-stopped` в docker-compose и systemd.

**Проблема 2: Логи занимают много места**

**Решение:** Используйте ротацию логов (logrotate).

**Проблема 3: Бот не подключается к БД**

**Решение:** Проверьте, что контейнеры в одной сети Docker и имена сервисов правильные.

---

## Что мы узнали

- Как создать Dockerfile для бота
- Как использовать Docker Compose
- Как настроить systemd service
- Как организовать мониторинг и логирование
- Как обновлять бота без простоя
- Как делать бэкапы
- Как настроить CI/CD

---

## Заключение

Поздравляем! Вы прошли весь путь от создания первого бота до его развёртывания в продакшене. Теперь у вас есть:

- Работающий Telegram-бот на Go
- Модульная архитектура
- База данных для хранения данных
- Админ-панель
- Система тестирования
- Готовое к продакшену приложение

**Следующие шаги:**
- Добавьте больше функций в бота
- Улучшите обработку ошибок
- Добавьте метрики и мониторинг
- Оптимизируйте производительность
- Расширьте тестовое покрытие

Удачи в разработке! 🚀

---

[Вернуться к оглавлению](./README.md)

