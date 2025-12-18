# Глава 6: Работа с базой данных

В этой главе мы научимся работать с базой данных PostgreSQL для хранения данных пользователей и состояния бота. Это позволит боту запоминать информацию между сеансами.

---

## 1. Зачем нужна база данных

**Проблема:**  
Без базы данных бот не может хранить данные между перезапусками. Вся информация теряется, когда программа завершается.

**Решение:**  
Используем PostgreSQL для хранения:
- Данных пользователей (ID, настройки, состояние)
- Истории сообщений (для анализа)
- Статистики бота
- Админских настроек

**Альтернативы:**
- SQLite — проще для начинающих, но менее производительна
- MongoDB — NoSQL база данных
- In-memory хранилище (Redis) — для кэширования

В этом руководстве используем PostgreSQL, так как это надёжная и популярная база данных.

---

## 2. Запускаем PostgreSQL через Docker

**Зачем Docker:**  
Docker позволяет запустить PostgreSQL одной командой, без установки на компьютер.

**Создаём файл `docker-compose.yml` в корне проекта:**
```yaml
version: '3.8'

services:
  # PostgreSQL — база данных
  postgres:
    image: postgres:15-alpine
    container_name: telegram-bot-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: telegram_bot
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**Разбор:**
- `image: postgres:15-alpine` — образ PostgreSQL версии 15 (alpine — лёгкая версия)
- `POSTGRES_DB: telegram_bot` — имя базы данных
- `ports: "5432:5432"` — пробрасываем порт из контейнера наружу
- `volumes` — сохраняем данные, чтобы они не пропали при перезапуске

**Запускаем:**
```bash
docker-compose up -d
```

**Проверяем:**
```bash
docker-compose ps
```

---

## 3. Создаём модели данных

**Что делаем:**  
Создаём Go-структуры, которые описывают данные пользователей.

**Создаём файл `internal/domain/user.go`:**
```go
package domain

import "time"

// User представляет пользователя бота
type User struct {
    ID           int64     `json:"id" db:"id"`                        // Telegram User ID
    Username     string    `json:"username" db:"username"`            // Username пользователя
    FirstName    string    `json:"first_name" db:"first_name"`        // Имя
    LastName     string    `json:"last_name" db:"last_name"`          // Фамилия
    LanguageCode string    `json:"language_code" db:"language_code"`  // Код языка
    IsActive     bool      `json:"is_active" db:"is_active"`          // Активен ли пользователь
    CreatedAt    time.Time `json:"created_at" db:"created_at"`        // Дата регистрации
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`        // Дата последнего обновления
}

// UserSettings настройки пользователя
type UserSettings struct {
    UserID          int64 `json:"user_id" db:"user_id"`
    NotificationsEnabled bool `json:"notifications_enabled" db:"notifications_enabled"` // Уведомления включены
    Language         string `json:"language" db:"language"`                            // Язык интерфейса
}
```

**Разбор:**

- `` `db:"id"` `` — тег для библиотеки работы с БД. Указывает, какое поле таблицы соответствует этому полю структуры.

- `int64` — используется для Telegram User ID, так как они большие числа.

- Вложенная структура `UserSettings` хранит настройки пользователя отдельно от основной информации.

---

## 4. Создаём SQL-миграции

**Что такое миграция:**  
Миграция — это SQL-скрипт, который изменяет структуру базы данных. Миграции позволяют версионировать изменения БД.

**Создаём файл `migrations/001_init.up.sql`:**
```sql
-- Создаём таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,                          -- Telegram User ID (первичный ключ)
    username VARCHAR(255),                          -- Username (может быть NULL)
    first_name VARCHAR(255) NOT NULL,               -- Имя (обязательное поле)
    last_name VARCHAR(255),                         -- Фамилия (может быть NULL)
    language_code VARCHAR(10) DEFAULT 'ru',         -- Код языка
    is_active BOOLEAN DEFAULT TRUE,                 -- Активен ли пользователь
    created_at TIMESTAMP DEFAULT NOW(),             -- Дата регистрации
    updated_at TIMESTAMP DEFAULT NOW()              -- Дата последнего обновления
);

-- Создаём таблицу настроек пользователей
CREATE TABLE IF NOT EXISTS user_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, -- Связь с таблицей users
    notifications_enabled BOOLEAN DEFAULT TRUE,     -- Уведомления включены
    language VARCHAR(10) DEFAULT 'ru'               -- Язык интерфейса
);

-- Создаём индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
```

**Разбор SQL:**

- `CREATE TABLE IF NOT EXISTS` — создать таблицу, если её ещё нет
- `BIGINT PRIMARY KEY` — большое целое число, первичный ключ
- `VARCHAR(255)` — строка до 255 символов
- `NOT NULL` — поле обязательное
- `DEFAULT` — значение по умолчанию
- `REFERENCES users(id)` — внешний ключ (ссылка на другую таблицу)
- `ON DELETE CASCADE` — при удалении пользователя удалятся его настройки
- `CREATE INDEX` — создаёт индекс для ускорения поиска

**Создаём файл отката `migrations/001_init.down.sql`:**
```sql
-- Удаляем таблицы в обратном порядке (из-за зависимостей)
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS users;
```

---

## 5. Подключаемся к базе данных

**Создаём файл `internal/repository/postgres.go`:**
```go
package repository

import (
    "database/sql"
    "fmt"

    // Импортируем драйвер PostgreSQL
    // Символ _ означает, что мы импортируем пакет только ради побочных эффектов
    _ "github.com/lib/pq"

    "telegram-bot/internal/config"
)

// PostgresDB — обёртка над подключением к PostgreSQL
type PostgresDB struct {
    DB *sql.DB // Стандартный интерфейс Go для работы с БД
}

// NewPostgresDB создаёт новое подключение к PostgreSQL
func NewPostgresDB(cfg config.DatabaseConfig) (*PostgresDB, error) {
    // Формируем строку подключения
    connStr := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host,
        cfg.Port,
        cfg.User,
        cfg.Password,
        cfg.Name,
        cfg.SSLMode,
    )

    // Открываем соединение с базой данных
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("ошибка открытия БД: %w", err)
    }

    // Проверяем, что соединение работает
    err = db.Ping()
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
    }

    return &PostgresDB{DB: db}, nil
}

// Close закрывает соединение с базой данных
func (p *PostgresDB) Close() error {
    return p.DB.Close()
}
```

**Разбор:**

- `_ "github.com/lib/pq"` — импорт с `_` означает, что мы не используем пакет напрямую. Драйвер регистрирует себя при импорте.

- `*sql.DB` — стандартный тип Go для работы с базами данных.

- `sql.Open("postgres", connStr)` — открывает соединение с БД.

- `db.Ping()` — проверяет, что соединение работает.

---

## 6. Создаём репозиторий для пользователей

**Что такое репозиторий:**  
Репозиторий — это слой, который отвечает за работу с базой данных. Он скрывает детали SQL от остального кода.

**Создаём файл `internal/repository/user_repository.go`:**
```go
package repository

import (
    "database/sql"
    "time"

    "telegram-bot/internal/domain"
)

// UserRepository — репозиторий для работы с пользователями
type UserRepository struct {
    db *sql.DB
}

// NewUserRepository создаёт новый репозиторий
func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// CreateOrUpdate создаёт пользователя или обновляет существующего
func (r *UserRepository) CreateOrUpdate(user *domain.User) error {
    query := `
        INSERT INTO users (id, username, first_name, last_name, language_code, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            language_code = EXCLUDED.language_code,
            updated_at = EXCLUDED.updated_at
    `

    now := time.Now()
    _, err := r.db.Exec(
        query,
        user.ID,
        user.Username,
        user.FirstName,
        user.LastName,
        user.LanguageCode,
        user.IsActive,
        now,
        now,
    )

    return err
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(id int64) (*domain.User, error) {
    query := `
        SELECT id, username, first_name, last_name, language_code, is_active, created_at, updated_at
        FROM users
        WHERE id = $1
    `

    user := &domain.User{}
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.FirstName,
        &user.LastName,
        &user.LanguageCode,
        &user.IsActive,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // Пользователь не найден
    }
    if err != nil {
        return nil, err
    }

    return user, nil
}

// GetAll возвращает всех активных пользователей
func (r *UserRepository) GetAll(limit, offset int) ([]*domain.User, error) {
    query := `
        SELECT id, username, first_name, last_name, language_code, is_active, created_at, updated_at
        FROM users
        WHERE is_active = true
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

    rows, err := r.db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*domain.User
    for rows.Next() {
        user := &domain.User{}
        err := rows.Scan(
            &user.ID,
            &user.Username,
            &user.FirstName,
            &user.LastName,
            &user.LanguageCode,
            &user.IsActive,
            &user.CreatedAt,
            &user.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, rows.Err()
}

// Count возвращает количество пользователей
func (r *UserRepository) Count() (int, error) {
    query := `SELECT COUNT(*) FROM users WHERE is_active = true`
    
    var count int
    err := r.db.QueryRow(query).Scan(&count)
    return count, err
}
```

**Разбор:**

- `ON CONFLICT (id) DO UPDATE` — если пользователь с таким ID уже есть, обновляем его данные.

- `QueryRow().Scan()` — выполняет запрос, который должен вернуть одну строку, и сканирует результаты в структуру.

- `rows.Next()` — перебирает строки результата запроса.

- `defer rows.Close()` — гарантирует закрытие результата запроса после выхода из функции.

---

## 7. Репозиторий для настроек пользователя

**Создаём файл `internal/repository/user_settings_repository.go`:**
```go
package repository

import (
    "database/sql"

    "telegram-bot/internal/domain"
)

// UserSettingsRepository — репозиторий для работы с настройками пользователей
type UserSettingsRepository struct {
    db *sql.DB
}

// NewUserSettingsRepository создаёт новый репозиторий
func NewUserSettingsRepository(db *sql.DB) *UserSettingsRepository {
    return &UserSettingsRepository{db: db}
}

// GetByUserID возвращает настройки пользователя
func (r *UserSettingsRepository) GetByUserID(userID int64) (*domain.UserSettings, error) {
    query := `
        SELECT user_id, notifications_enabled, language
        FROM user_settings
        WHERE user_id = $1
    `

    settings := &domain.UserSettings{}
    err := r.db.QueryRow(query, userID).Scan(
        &settings.UserID,
        &settings.NotificationsEnabled,
        &settings.Language,
    )

    if err == sql.ErrNoRows {
        // Настроек нет — возвращаем настройки по умолчанию
        return &domain.UserSettings{
            UserID:              userID,
            NotificationsEnabled: true,
            Language:            "ru",
        }, nil
    }

    return settings, err
}

// Update обновляет настройки пользователя
func (r *UserSettingsRepository) Update(settings *domain.UserSettings) error {
    query := `
        INSERT INTO user_settings (user_id, notifications_enabled, language)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE SET
            notifications_enabled = EXCLUDED.notifications_enabled,
            language = EXCLUDED.language
    `

    _, err := r.db.Exec(query, settings.UserID, settings.NotificationsEnabled, settings.Language)
    return err
}
```

---

## 8. Интегрируем базу данных в бота

**Обновляем `cmd/bot/main.go`:**
```go
package main

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/config"
    "telegram-bot/internal/handler"
    "telegram-bot/internal/middleware"
    "telegram-bot/internal/repository"
)

func main() {
    // Загружаем конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Ошибка загрузки конфигурации:", err)
    }

    // Подключаемся к базе данных
    postgresDB, err := repository.NewPostgresDB(cfg.Database)
    if err != nil {
        log.Fatal("Ошибка подключения к БД:", err)
    }
    defer postgresDB.Close()

    // Создаём репозитории
    userRepo := repository.NewUserRepository(postgresDB.DB)
    settingsRepo := repository.NewUserSettingsRepository(postgresDB.DB)

    // Создаём экземпляр бота
    bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
    if err != nil {
        log.Fatal("Ошибка создания бота:", err)
    }

    bot.Debug = cfg.Bot.Debug
    log.Printf("Авторизован как %s", bot.Self.UserName)

    // Создаём диспетчер обработчиков
    dispatcher := handler.NewDispatcher()
    dispatcher.Register(handler.NewStartHandler())
    dispatcher.Register(handler.NewHelpHandler())
    dispatcher.Register(handler.NewInfoHandler())

    // Создаём обработчик сообщений с репозиториями
    messageHandler := handler.NewMessageHandler()

    // Настраиваем получение обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = cfg.Bot.Timeout
    updates := bot.GetUpdatesChan(u)

    // Обрабатываем обновления
    for update := range updates {
        handleUpdate(bot, dispatcher, messageHandler, userRepo, update)
    }
}

import (
    "telegram-bot/internal/domain"
    // ... остальные импорты
)

func handleUpdate(
    bot *tgbotapi.BotAPI,
    dispatcher *handler.Dispatcher,
    messageHandler *handler.MessageHandler,
    userRepo *repository.UserRepository,
    update tgbotapi.Update,
) {
    if update.Message == nil {
        return
    }

    msg := update.Message

    // Сохраняем или обновляем пользователя в БД
    user := &domain.User{
        ID:           int64(msg.From.ID),
        Username:     msg.From.UserName,
        FirstName:    msg.From.FirstName,
        LastName:     msg.From.LastName,
        LanguageCode: msg.From.LanguageCode,
        IsActive:     true,
    }
    
    if err := userRepo.CreateOrUpdate(user); err != nil {
        log.Printf("Ошибка сохранения пользователя: %v", err)
    }

    if msg.IsCommand() {
        middleware.LogCommand(msg)
        err := dispatcher.HandleCommand(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки команды: %v", err)
        }
        return
    }

    if msg.Text != "" {
        middleware.LogMessage(msg)
        err := messageHandler.Handle(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки сообщения: %v", err)
        }
    }
}
```

---

## Типичные ошибки

**Ошибка 1: "connection refused"**

**Причина:** PostgreSQL не запущен или неправильный адрес/порт.

**Решение:** 
1. Проверьте, что Docker-контейнер запущен: `docker-compose ps`
2. Проверьте настройки подключения в `.env`

**Ошибка 2: "relation does not exist"**

**Причина:** Таблицы не созданы в базе данных.

**Решение:** Выполните миграции SQL вручную или используйте инструмент для миграций.

**Ошибка 3: "sql: connection is already closed"**

**Причина:** Попытка использовать закрытое соединение.

**Решение:** Убедитесь, что `defer postgresDB.Close()` находится после успешного подключения.

---

## Что мы узнали

- Как запустить PostgreSQL через Docker
- Как создать модели данных для пользователей
- Как создать SQL-миграции
- Как подключиться к базе данных из Go
- Как создать репозитории для работы с данными
- Как интегрировать БД в бота
- Типичные ошибки и их решение

---

[Следующая глава: Создание админ-панели](./07-admin-panel.md)

[Вернуться к оглавлению](./README.md)

