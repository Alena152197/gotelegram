# Глава 8: Дополнительные функции

В этой главе мы добавим продвинутые функции в бота: работу с файлами, интеграцию с внешними API, планировщик задач и другие полезные возможности.

---

## 1. Работа с файлами

**Что делаем:**  
Научим бота отправлять и получать файлы (документы, фото, видео).

### Отправка файлов

**Добавляем функцию отправки документа:**
```go
// sendDocument отправляет документ пользователю
func sendDocument(bot *tgbotapi.BotAPI, chatID int64, filePath string, caption string) error {
    file := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
    file.Caption = caption
    _, err := bot.Send(file)
    return err
}

// sendPhoto отправляет фото
func sendPhoto(bot *tgbotapi.BotAPI, chatID int64, filePath string, caption string) error {
    photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(filePath))
    photo.Caption = caption
    _, err := bot.Send(photo)
    return err
}
```

### Получение файлов

**Создаём обработчик для загрузки файлов:**
```go
// handleDocument обрабатывает получение документа
func handleDocument(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    doc := msg.Document
    chatID := msg.Chat.ID

    // Получаем информацию о файле
    fileInfo := fmt.Sprintf(
        "📎 Получен файл:\n"+
        "Имя: %s\n"+
        "Размер: %.2f KB\n"+
        "Тип: %s",
        doc.FileName,
        float64(doc.FileSize)/1024,
        doc.MimeType,
    )

    reply := tgbotapi.NewMessage(chatID, fileInfo)
    bot.Send(reply)

    // Скачиваем файл
    fileConfig := tgbotapi.FileConfig{FileID: doc.FileID}
    file, err := bot.GetFile(fileConfig)
    if err != nil {
        log.Printf("Ошибка получения файла: %v", err)
        return
    }

    // Сохраняем файл
    // TODO: Реализовать сохранение файла
    log.Printf("Файл получен: %s", file.FilePath)
}
```

---

## 2. Работа с голосовыми сообщениями

**Добавляем обработку голосовых сообщений:**
```go
// handleVoice обрабатывает голосовое сообщение
func handleVoice(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    voice := msg.Voice
    chatID := msg.Chat.ID

    duration := voice.Duration
    reply := tgbotapi.NewMessage(chatID, 
        fmt.Sprintf("🎤 Получено голосовое сообщение длительностью %d секунд", duration))
    bot.Send(reply)
}
```

---

## 3. Интеграция с внешними API

**Пример: получение курса валют**

**Создаём файл `internal/service/exchange_service.go`:**
```go
package service

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

// ExchangeService получает курсы валют
type ExchangeService struct {
    apiURL string
}

// NewExchangeService создаёт новый сервис
func NewExchangeService() *ExchangeService {
    return &ExchangeService{
        apiURL: "https://api.exchangerate-api.com/v4/latest/USD",
    }
}

// GetRate получает курс валюты
func (s *ExchangeService) GetRate(currency string) (float64, error) {
    resp, err := http.Get(s.apiURL)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        return 0, err
    }

    rates := data["rates"].(map[string]interface{})
    rate, ok := rates[currency].(float64)
    if !ok {
        return 0, fmt.Errorf("валюта не найдена")
    }

    return rate, nil
}
```

**Создаём обработчик команды /rate:**
```go
// RateHandler обрабатывает команду /rate
type RateHandler struct {
    exchangeService *service.ExchangeService
}

func (h *RateHandler) Command() string {
    return "rate"
}

func (h *RateHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID

    // Получаем курс рубля
    rate, err := h.exchangeService.GetRate("RUB")
    if err != nil {
        reply := tgbotapi.NewMessage(chatID, "❌ Ошибка получения курса валют.")
        bot.Send(reply)
        return err
    }

    text := fmt.Sprintf("💰 Курс USD/RUB: %.2f", rate)
    reply := tgbotapi.NewMessage(chatID, text)
    bot.Send(reply)

    return nil
}
```

---

## 4. Планировщик задач

**Что делаем:**  
Создаём систему для выполнения задач по расписанию (например, ежедневная рассылка).

**Устанавливаем библиотеку:**
```bash
go get github.com/robfig/cron/v3
```

**Создаём файл `internal/service/scheduler.go`:**
```go
package service

import (
    "log"

    "github.com/robfig/cron/v3"
)

// Scheduler управляет запланированными задачами
type Scheduler struct {
    cron *cron.Cron
}

// NewScheduler создаёт новый планировщик
func NewScheduler() *Scheduler {
    return &Scheduler{
        cron: cron.New(),
    }
}

// Start запускает планировщик
func (s *Scheduler) Start() {
    s.cron.Start()
    log.Println("Планировщик задач запущен")
}

// Stop останавливает планировщик
func (s *Scheduler) Stop() {
    s.cron.Stop()
    log.Println("Планировщик задач остановлен")
}

// AddDailyTask добавляет задачу, выполняемую ежедневно
func (s *Scheduler) AddDailyTask(schedule string, task func()) (cron.EntryID, error) {
    return s.cron.AddFunc(schedule, task)
}

// Пример использования:
// scheduler.AddDailyTask("0 9 * * *", func() {
//     // Задача выполнится каждый день в 9:00
// })
```

---

## 5. Кэширование с Redis

**Зачем нужно:**  
Кэширование позволяет ускорить работу бота, храня часто используемые данные в памяти.

**Устанавливаем библиотеку:**
```bash
go get github.com/redis/go-redis/v9
```

**Создаём файл `internal/repository/cache.go`:**
```go
package repository

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/redis/go-redis/v9"
)

// Cache предоставляет функции кэширования
type Cache struct {
    client *redis.Client
    ctx    context.Context
}

// NewCache создаёт новый кэш
func NewCache(addr string) *Cache {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })

    return &Cache{
        client: client,
        ctx:    context.Background(),
    }
}

// Set сохраняет значение в кэш
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Get получает значение из кэша
func (c *Cache) Get(key string, dest interface{}) error {
    data, err := c.client.Get(c.ctx, key).Result()
    if err == redis.Nil {
        return ErrCacheMiss // Ключ не найден
    }
    if err != nil {
        return err
    }

    return json.Unmarshal([]byte(data), dest)
}

import (
    "errors"
    // ... остальные импорты
)

var ErrCacheMiss = errors.New("ключ не найден в кэше")
```

---

## 6. Обработка ошибок и логирование

**Создаём структурированное логирование:**
```go
// setupLogger настраивает логирование
func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
    var logger *zap.Logger
    var err error

    if cfg.Level == "debug" {
        logger, err = zap.NewDevelopment()
    } else {
        logger, err = zap.NewProduction()
    }

    if err != nil {
        return nil, err
    }

    return logger, nil
}

// Использование:
// logger.Info("Бот запущен",
//     zap.String("username", bot.Self.UserName),
//     zap.Int64("id", bot.Self.ID),
// )
```

---

## Типичные ошибки

**Ошибка 1: "file too large" при отправке файла**

**Причина:** Telegram ограничивает размер файлов (обычно 20 MB для ботов).

**Решение:** Проверяйте размер файла перед отправкой и сжимайте при необходимости.

**Ошибка 2: Timeout при запросе к внешнему API**

**Причина:** Внешний сервис не отвечает вовремя.

**Решение:** Используйте контексты с таймаутом для HTTP-запросов.

---

## Что мы узнали

- Как работать с файлами (отправка и получение)
- Как обрабатывать голосовые сообщения
- Как интегрироваться с внешними API
- Как создать планировщик задач
- Как использовать Redis для кэширования
- Как настроить структурированное логирование

---

[Следующая глава: Тестирование и отладка](./09-testing-debugging.md)

[Вернуться к оглавлению](./README.md)

