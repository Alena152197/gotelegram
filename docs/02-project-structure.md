# Глава 2: Структура проекта и конфигурация

В этой главе мы организуем структуру папок проекта и научимся работать с конфигурацией через переменные окружения. Для телеграм-бота это особенно важно, так как нужно безопасно хранить токен бота.

---

## 1. Создаём структуру папок

**Зачем это нужно:**  
Правильная организация кода помогает не запутаться в проекте. Каждая папка отвечает за свою часть функционала.

**Что делаем:**  
Создаём папки согласно стандартной структуре Go-проектов.

**Команды:**
```bash
# Создаём основные папки
mkdir -p cmd/bot
mkdir -p internal/config
mkdir -p internal/domain
mkdir -p internal/handler
mkdir -p internal/service
mkdir -p internal/repository
mkdir -p internal/middleware
mkdir -p internal/keyboard
mkdir -p migrations
mkdir -p pkg/logger
```

**Что означает каждая папка:**

| Папка | Назначение |
|-------|------------|
| `cmd/bot` | Точка входа для бота (файл main.go) |
| `internal/config` | Загрузка конфигурации из переменных окружения |
| `internal/domain` | Модели данных (структуры) |
| `internal/handler` | Обработчики команд и сообщений |
| `internal/service` | Бизнес-логика бота |
| `internal/repository` | Работа с базой данных |
| `internal/middleware` | Промежуточное ПО (логирование, проверка прав) |
| `internal/keyboard` | Клавиатуры и кнопки |
| `migrations` | SQL-миграции для базы данных |
| `pkg/logger` | Утилиты для логирования |

**Что такое `internal`:**  
Папка `internal` — специальная в Go. Код внутри неё можно использовать только в этом проекте. Другие проекты не смогут импортировать эти пакеты.

**Что такое `cmd`:**  
Папка `cmd` содержит точки входа в приложение. Каждая подпапка — отдельная программа, которую можно запустить.

---

## 2. Создаём файл конфигурации

**Зачем это нужно:**  
Конфигурация (токен бота, пароли, адреса баз данных) не должна быть "зашита" в код. Её нужно хранить отдельно, чтобы легко менять без перекомпиляции. Для токена бота это критически важно — никогда не коммитьте токен в Git!

**Создаём файл `.env` в корне проекта:**
```
# Telegram Bot Token
# Получите токен у @BotFather в Telegram
BOT_TOKEN=your_bot_token_here

# База данных
DB_HOST=localhost
DB_PORT=5432
DB_NAME=telegram_bot
DB_USER=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable

# Настройки бота
BOT_DEBUG=false
BOT_TIMEOUT=60

# Админы бота (через запятую, Telegram User ID)
ADMIN_IDS=123456789,987654321

# Настройки логирования
LOG_LEVEL=info
LOG_FILE=bot.log
```

**Разбор:**
- Каждая строка — это пара "ключ=значение"
- `#` — начало комментария
- Пробелов вокруг `=` быть не должно
- Значения без кавычек

**Важно:**  
Добавьте `.env` в файл `.gitignore`, чтобы случайно не выложить токен в Git:

**Создаём файл `.gitignore`:**
```
# Переменные окружения
.env
.env.local

# Бинарные файлы
*.exe
*.exe~
*.dll
*.so
*.dylib

# Тестовые бинарные файлы
*.test

# Выходные файлы
*.out

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Логи
*.log

# База данных (если используете SQLite)
*.db
*.sqlite
*.sqlite3
```

---

## 3. Устанавливаем библиотеки

**Что делаем:**  
Устанавливаем библиотеки для работы с конфигурацией, Telegram Bot API и другие зависимости проекта.

**Команды:**
```bash
# Библиотека для чтения .env файлов
go get github.com/joho/godotenv

# Библиотека для работы с конфигурацией
go get github.com/kelseyhightower/envconfig

# Telegram Bot API библиотека
go get github.com/go-telegram-bot-api/telegram-bot-api/v5

# Драйвер PostgreSQL (будем использовать позже)
go get github.com/lib/pq

# Библиотека для работы с UUID
go get github.com/google/uuid

# Библиотека для логирования
go get go.uber.org/zap
```

**Что такое `go get`:**  
Команда `go get` скачивает внешнюю библиотеку и добавляет её в файл `go.mod`. После этого библиотеку можно использовать в коде.

**Результат:**  
В файле `go.mod` появятся строки с зависимостями, а также создастся файл `go.sum` с контрольными суммами (для безопасности).

---

## 4. Создаём структуру конфигурации

**Что делаем:**  
Создаём Go-структуру, которая будет хранить все настройки приложения.

**Создаём файл `internal/config/config.go`:**
```go
package config

import (
    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

// Config — главная структура конфигурации приложения
// Все поля заполняются из переменных окружения
type Config struct {
    Bot      BotConfig      // Настройки бота
    Database DatabaseConfig // Настройки базы данных
    Logging  LoggingConfig  // Настройки логирования
}

// BotConfig — настройки Telegram-бота
type BotConfig struct {
    Token    string   `envconfig:"BOT_TOKEN" required:"true"`          // Токен бота (обязательный)
    Debug    bool     `envconfig:"BOT_DEBUG" default:"false"`          // Режим отладки
    Timeout  int      `envconfig:"BOT_TIMEOUT" default:"60"`           // Таймаут запросов (секунды)
    AdminIDs []int64  `envconfig:"ADMIN_IDS"`                          // ID администраторов
}

// DatabaseConfig — настройки подключения к PostgreSQL
type DatabaseConfig struct {
    Host     string `envconfig:"DB_HOST" default:"localhost"`     // Адрес сервера БД
    Port     int    `envconfig:"DB_PORT" default:"5432"`          // Порт БД
    Name     string `envconfig:"DB_NAME" default:"telegram_bot"`  // Имя базы данных
    User     string `envconfig:"DB_USER" default:"postgres"`      // Пользователь БД
    Password string `envconfig:"DB_PASSWORD" default:""`          // Пароль БД
    SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`   // Режим SSL
}

// LoggingConfig — настройки логирования
type LoggingConfig struct {
    Level string `envconfig:"LOG_LEVEL" default:"info"`  // Уровень логирования (debug, info, warn, error)
    File  string `envconfig:"LOG_FILE" default:"bot.log"` // Файл для логов
}
```

**Разбор:**

- `package config` — объявляем, что этот файл принадлежит пакету `config`. Имя пакета обычно совпадает с именем папки.

- `import (...)` — импортируем несколько пакетов. Когда импортов много, их группируют в скобках.

- `type Config struct {...}` — создаём структуру `Config`, которая содержит другие структуры (вложенные структуры).

- `` `envconfig:"BOT_TOKEN" required:"true"` `` — это **теги структуры**. Они не влияют на работу Go, но библиотеки могут их читать:
  - `envconfig:"BOT_TOKEN"` — библиотека envconfig будет искать переменную окружения `BOT_TOKEN`
  - `required:"true"` — поле обязательное, без него программа не запустится
  - `default:"false"` — если переменная не найдена, использовать значение по умолчанию

- `[]int64` — срез (список) чисел типа `int64`. Используется для хранения списка ID администраторов.

---

## 5. Функция загрузки конфигурации

**Что делаем:**  
Добавляем функцию, которая читает переменные окружения и заполняет структуру `Config`.

**Добавляем в `internal/config/config.go`:**
```go
// Load загружает конфигурацию из переменных окружения
// Сначала пытается прочитать файл .env, затем читает переменные окружения
func Load() (*Config, error) {
    // Пытаемся загрузить .env файл
    // Если файла нет — не страшно, будем читать из системных переменных
    _ = godotenv.Load()

    // Создаём пустую структуру конфигурации
    var cfg Config

    // Заполняем структуру из переменных окружения
    // Если обязательное поле отсутствует — вернётся ошибка
    err := envconfig.Process("", &cfg)
    if err != nil {
        return nil, err
    }

    // Парсим список ID администраторов из строки
    // ADMIN_IDS="123456789,987654321" -> []int64{123456789, 987654321}
    if err := parseAdminIDs(&cfg); err != nil {
        return nil, err
    }

    // Возвращаем указатель на конфигурацию
    return &cfg, nil
}
```

**Разбор:**

- `func Load() (*Config, error)` — функция возвращает два значения:
  - `*Config` — указатель на структуру Config
  - `error` — ошибка (если что-то пошло не так)

- `_ = godotenv.Load()` — вызываем функцию загрузки .env файла
  - `_` (нижнее подчёркивание) означает "игнорируем возвращаемое значение". Мы не проверяем ошибку, потому что файл .env необязателен.

- `var cfg Config` — создаём переменную `cfg` типа `Config`. Все поля будут иметь нулевые значения.

- `envconfig.Process("", &cfg)` — заполняем структуру из переменных окружения
  - `""` — пустой префикс (переменные читаются как есть, без префикса)
  - `&cfg` — передаём **адрес** переменной (указатель). Символ `&` означает "взять адрес". Это нужно, чтобы функция могла изменить нашу переменную.

- `if err != nil` — проверяем, была ли ошибка
  - `nil` — специальное значение "ничего" для указателей, интерфейсов и ошибок
  - Если `err` не `nil`, значит произошла ошибка

- `return &cfg, nil` — возвращаем указатель на конфигурацию и `nil` вместо ошибки (ошибки не было)

**Что такое указатель:**  
Указатель — это адрес в памяти, где хранится значение. Вместо копирования всей структуры, мы передаём только её адрес. Это экономит память и позволяет изменять оригинал.
- `&cfg` — получить адрес переменной `cfg`
- `*Config` — тип "указатель на Config"

---

## 6. Парсинг ID администраторов

**Что делаем:**  
Добавляем функцию для парсинга строки с ID администраторов (например, "123,456,789") в срез чисел.

**Добавляем вспомогательные функции в `internal/config/config.go`:**
```go
import (
    "strconv"
    "strings"

    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

// parseAdminIDs парсит строку ADMIN_IDS и заполняет BotConfig.AdminIDs
func parseAdminIDs(cfg *Config) error {
    // Получаем значение переменной окружения ADMIN_IDS
    adminIDsStr := os.Getenv("ADMIN_IDS")
    
    // Если переменная не задана — возвращаем nil (это не ошибка)
    if adminIDsStr == "" {
        cfg.Bot.AdminIDs = []int64{}
        return nil
    }

    // Разделяем строку по запятым: "123,456,789" -> ["123", "456", "789"]
    parts := strings.Split(adminIDsStr, ",")
    
    // Создаём срез для хранения ID
    adminIDs := make([]int64, 0, len(parts))
    
    // Перебираем каждую часть
    for _, part := range parts {
        // Убираем пробелы с начала и конца
        part = strings.TrimSpace(part)
        
        // Пропускаем пустые строки
        if part == "" {
            continue
        }
        
        // Преобразуем строку в число (int64)
        id, err := strconv.ParseInt(part, 10, 64)
        if err != nil {
            return err
        }
        
        // Добавляем ID в срез
        adminIDs = append(adminIDs, id)
    }
    
    // Сохраняем результат в конфигурацию
    cfg.Bot.AdminIDs = adminIDs
    
    return nil
}
```

**Разбор:**

- `os.Getenv("ADMIN_IDS")` — получает значение переменной окружения. Если переменной нет, возвращает пустую строку.

- `strings.Split(adminIDsStr, ",")` — разделяет строку по запятой и возвращает срез строк.

- `strings.TrimSpace(part)` — убирает пробелы в начале и конце строки.

- `strconv.ParseInt(part, 10, 64)` — преобразует строку в число
  - `part` — строка для преобразования
  - `10` — основание системы счисления (десятичная)
  - `64` — размер числа в битах (64 бита = int64)

- `make([]int64, 0, len(parts))` — создаёт срез с начальной длиной 0 и ёмкостью len(parts). Это оптимизация — мы заранее резервируем место.

---

## 7. Полный код файла config.go

**Создаём файл `internal/config/config.go`:**
```go
package config

import (
    "os"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

// Config — главная структура конфигурации приложения
type Config struct {
    Bot      BotConfig
    Database DatabaseConfig
    Logging  LoggingConfig
}

// BotConfig — настройки Telegram-бота
type BotConfig struct {
    Token    string  `envconfig:"BOT_TOKEN" required:"true"`
    Debug    bool    `envconfig:"BOT_DEBUG" default:"false"`
    Timeout  int     `envconfig:"BOT_TIMEOUT" default:"60"`
    AdminIDs []int64 `envconfig:"-"` // Игнорируем envconfig, парсим вручную
}

// DatabaseConfig — настройки подключения к PostgreSQL
type DatabaseConfig struct {
    Host     string `envconfig:"DB_HOST" default:"localhost"`
    Port     int    `envconfig:"DB_PORT" default:"5432"`
    Name     string `envconfig:"DB_NAME" default:"telegram_bot"`
    User     string `envconfig:"DB_USER" default:"postgres"`
    Password string `envconfig:"DB_PASSWORD" default:""`
    SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`
}

// LoggingConfig — настройки логирования
type LoggingConfig struct {
    Level string `envconfig:"LOG_LEVEL" default:"info"`
    File  string `envconfig:"LOG_FILE" default:"bot.log"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
    // Загружаем .env файл (если есть)
    _ = godotenv.Load()

    // Создаём структуру конфигурации
    var cfg Config

    // Заполняем из переменных окружения
    err := envconfig.Process("", &cfg)
    if err != nil {
        return nil, err
    }

    // Парсим список ID администраторов
    if err := parseAdminIDs(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// parseAdminIDs парсит строку ADMIN_IDS и заполняет BotConfig.AdminIDs
func parseAdminIDs(cfg *Config) error {
    adminIDsStr := os.Getenv("ADMIN_IDS")
    
    if adminIDsStr == "" {
        cfg.Bot.AdminIDs = []int64{}
        return nil
    }

    parts := strings.Split(adminIDsStr, ",")
    adminIDs := make([]int64, 0, len(parts))
    
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        
        id, err := strconv.ParseInt(part, 10, 64)
        if err != nil {
            return err
        }
        
        adminIDs = append(adminIDs, id)
    }
    
    cfg.Bot.AdminIDs = adminIDs
    return nil
}
```

---

## 8. Создаём точку входа бота

**Что делаем:**  
Создаём главный файл бота, который загружает конфигурацию и выводит информацию.

**Создаём файл `cmd/bot/main.go`:**
```go
package main

import (
    "fmt"
    "log"

    "telegram-bot/internal/config"
)

func main() {
    // Загружаем конфигурацию
    cfg, err := config.Load()
    
    // Проверяем ошибку загрузки
    if err != nil {
        // log.Fatal выводит сообщение и завершает программу
        log.Fatal("Ошибка загрузки конфигурации:", err)
    }

    // Выводим информацию о конфигурации (без токена!)
    fmt.Println("=== Telegram Bot ===")
    fmt.Printf("Режим отладки: %v\n", cfg.Bot.Debug)
    fmt.Printf("Таймаут: %d сек\n", cfg.Bot.Timeout)
    fmt.Printf("Админов: %d\n", len(cfg.Bot.AdminIDs))
    fmt.Printf("База данных: %s@%s:%d/%s\n", 
        cfg.Database.User, 
        cfg.Database.Host, 
        cfg.Database.Port, 
        cfg.Database.Name,
    )
    fmt.Printf("Уровень логирования: %s\n", cfg.Logging.Level)
    
    // TODO: Здесь будет запуск бота
    fmt.Println("\nБот пока не реализован...")
}
```

**Разбор:**

- `import "telegram-bot/internal/config"` — импортируем наш пакет config. Путь начинается с имени модуля (`telegram-bot`), указанного в `go.mod`.

- `log.Fatal(...)` — выводит сообщение и завершает программу с кодом ошибки 1. Используется для критических ошибок, после которых продолжение невозможно.

- `fmt.Printf(...)` — форматированный вывод:
  - `%v` — значение в формате по умолчанию
  - `%d` — целое число
  - `%s` — строка
  - `\n` — перенос строки

- `cfg.Bot.Debug` — обращаемся к вложенной структуре через точку

**Важно:**  
Никогда не выводите токен бота в логи или консоль! Это критическая информация безопасности.

---

## 9. Проверяем работу

**Запускаем программу:**
```bash
go run cmd/bot/main.go
```

**Ожидаемый результат (если токен не задан):**
```
Ошибка загрузки конфигурации: required key BOT_TOKEN missing value
```

Это нормально — мы ещё не задали токен. Сейчас нам важно проверить, что конфигурация загружается правильно.

**Создайте файл `.env` с тестовыми данными:**
```
BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
BOT_DEBUG=true
BOT_TIMEOUT=60
ADMIN_IDS=123456789
DB_HOST=localhost
DB_PORT=5432
DB_NAME=telegram_bot
DB_USER=postgres
DB_PASSWORD=secret
LOG_LEVEL=info
LOG_FILE=bot.log
```

**Запускаем снова:**
```bash
go run cmd/bot/main.go
```

**Ожидаемый результат:**
```
=== Telegram Bot ===
Режим отладки: true
Таймаут: 60 сек
Админов: 1
База данных: postgres@localhost:5432/telegram_bot
Уровень логирования: info

Бот пока не реализован...
```

---

## 10. Структура проекта

**Как должна выглядеть структура проекта сейчас:**

```
telegram-bot/
├── cmd/
│   └── bot/
│       └── main.go
├── internal/
│   └── config/
│       └── config.go
├── migrations/
├── pkg/
│   └── logger/
├── .env
├── .gitignore
├── go.mod
└── go.sum
```

---

## Типичные ошибки

**Ошибка 1: "required key BOT_TOKEN missing value"**

**Причина:** Переменная `BOT_TOKEN` не задана в `.env` файле или в переменных окружения.

**Решение:** 
1. Проверьте, что файл `.env` существует в корне проекта
2. Убедитесь, что в файле есть строка `BOT_TOKEN=your_token_here`
3. Проверьте, что нет пробелов вокруг знака `=`

**Ошибка 2: "cannot find package"**

**Причина:** Путь импорта не совпадает с именем модуля в `go.mod`.

**Решение:**
1. Откройте файл `go.mod` и проверьте первую строку (например, `module telegram-bot`)
2. Убедитесь, что импорты используют это же имя: `import "telegram-bot/internal/config"`

**Ошибка 3: "no required module provides package"**

**Причина:** Зависимости не установлены.

**Решение:**
```bash
go mod download
go mod tidy
```

---

## Словарь терминов

| Термин | Объяснение |
|--------|------------|
| **Пакет (package)** | Способ организации кода в Go. Каждый файл принадлежит какому-то пакету |
| **Модуль (module)** | Проект Go с файлом go.mod. Содержит один или несколько пакетов |
| **Указатель** | Адрес в памяти, где хранится значение. Обозначается `*` |
| **Теги структуры** | Метаданные полей структуры в обратных кавычках. Читаются библиотеками |
| **nil** | Специальное значение "ничего" для указателей, интерфейсов, ошибок |
| **Переменные окружения** | Системные переменные, доступные всем программам |

---

## Что мы узнали

- Как организовать структуру Go-проекта
- Зачем нужны папки `cmd`, `internal`, `pkg`
- Как хранить конфигурацию в файле `.env`
- Как создавать структуры с тегами
- Как загружать конфигурацию из переменных окружения
- Что такое указатели и зачем они нужны
- Как обрабатывать ошибки в Go
- Как парсить сложные данные (списки ID)

---

[Следующая глава: Работа с Telegram Bot API](./03-telegram-api.md)

[Вернуться к оглавлению](./README.md)

