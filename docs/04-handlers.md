# Глава 4: Обработка команд и сообщений

В этой главе мы реорганизуем код бота, создав модульную систему обработчиков. Это сделает код более структурированным и удобным для расширения.

---

## 1. Зачем нужна модульная структура

**Проблема текущего кода:**  
Весь код обработки находится в одном файле `main.go`. Когда команд станет много, файл станет очень большим и неудобным для работы.

**Решение:**  
Разделим код на модули:
- Отдельные файлы для каждого типа обработчиков
- Структура для хранения состояния бота
- Middleware для общих задач (логирование, проверка прав)

**Преимущества:**
- Код проще читать и поддерживать
- Легко добавлять новые команды
- Можно переиспользовать общую логику

---

## 2. Создаём структуру Bot

**Что делаем:**  
Создаём структуру, которая будет хранить состояние бота и все необходимые зависимости.

**Создаём файл `internal/domain/bot.go`:**
```go
package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Bot представляет экземпляр бота со всеми зависимостями
type Bot struct {
    API    *tgbotapi.BotAPI  // API для работы с Telegram
    Config *BotConfig         // Конфигурация бота
}

// BotConfig содержит настройки бота
type BotConfig struct {
    Token    string
    Debug    bool
    AdminIDs []int64
}
```

**Разбор:**

- `domain` — пакет для доменных моделей (структур данных). Это стандартное название для такой папки.

- Структура `Bot` хранит всё, что нужно для работы бота:
  - `API` — экземпляр Telegram Bot API
  - `Config` — конфигурация

- В будущем сюда можно добавить репозитории для работы с базой данных, сервисы и т.д.

---

## 3. Создаём базовый обработчик

**Что делаем:**  
Создаём интерфейс для обработчиков команд, чтобы все команды работали одинаково.

**Создаём файл `internal/handler/handler.go`:**
```go
package handler

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Handler — интерфейс для обработчиков команд
type Handler interface {
    Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error
    Command() string  // Возвращает команду, которую обрабатывает этот обработчик
}
```

**Разбор:**

- **Интерфейс в Go** — это набор методов, которые тип должен реализовать. Если структура имеет все методы интерфейса, она автоматически его реализует.

- `Handler` — интерфейс с двумя методами:
  - `Handle` — обрабатывает сообщение
  - `Command` — возвращает имя команды

- Любая структура, которая реализует эти методы, может использоваться как обработчик.

---

## 4. Создаём обработчик команды /start

**Что делаем:**  
Создаём отдельный обработчик для команды `/start`.

**Создаём файл `internal/handler/start_handler.go`:**
```go
package handler

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartHandler обрабатывает команду /start
type StartHandler struct{}

// NewStartHandler создаёт новый обработчик команды /start
func NewStartHandler() *StartHandler {
    return &StartHandler{}
}

// Command возвращает команду, которую обрабатывает этот обработчик
func (h *StartHandler) Command() string {
    return "start"
}

// Handle обрабатывает команду /start
func (h *StartHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID

    text := "Привет! Я тестовый бот на Go.\n\n" +
        "Я могу помочь вам с различными задачами.\n\n" +
        "Доступные команды:\n" +
        "/start - начать работу\n" +
        "/help - помощь\n" +
        "/info - информация о вас"

    reply := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(reply)
    return err
}
```

**Разбор:**

- `StartHandler` — структура, реализующая интерфейс `Handler`.

- `NewStartHandler()` — функция-конструктор. По соглашению Go, функции, создающие новые экземпляры, начинаются с `New`.

- `func (h *StartHandler) Command()` — метод структуры. `(h *StartHandler)` означает, что метод принадлежит структуре `StartHandler`.

- `return err` — возвращаем ошибку, если отправка не удалась. Вызывающий код может её обработать.

---

## 5. Создаём обработчик команды /help

**Создаём файл `internal/handler/help_handler.go`:**
```go
package handler

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler обрабатывает команду /help
type HelpHandler struct{}

// NewHelpHandler создаёт новый обработчик команды /help
func NewHelpHandler() *HelpHandler {
    return &HelpHandler{}
}

// Command возвращает команду
func (h *HelpHandler) Command() string {
    return "help"
}

// Handle обрабатывает команду /help
func (h *HelpHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID

    text := "Это справочная информация.\n\n" +
        "<b>Доступные команды:</b>\n\n" +
        "/start - начать работу с ботом\n" +
        "/help - показать эту справку\n" +
        "/info - информация о вашем профиле\n\n" +
        "Бот создан с помощью библиотеки go-telegram-bot-api."

    reply := tgbotapi.NewMessage(chatID, text)
    reply.ParseMode = tgbotapi.ModeHTML
    _, err := bot.Send(reply)
    return err
}
```

---

## 6. Создаём обработчик команды /info

**Создаём файл `internal/handler/info_handler.go`:**
```go
package handler

import (
    "fmt"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InfoHandler обрабатывает команду /info
type InfoHandler struct{}

// NewInfoHandler создаёт новый обработчик команды /info
func NewInfoHandler() *InfoHandler {
    return &InfoHandler{}
}

// Command возвращает команду
func (h *InfoHandler) Command() string {
    return "info"
}

// Handle обрабатывает команду /info
func (h *InfoHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID
    user := msg.From

    info := "<b>Информация о вас:</b>\n\n"
    info += fmt.Sprintf("<b>ID:</b> <code>%d</code>\n", user.ID)
    info += fmt.Sprintf("<b>Имя:</b> %s\n", user.FirstName)

    if user.LastName != "" {
        info += fmt.Sprintf("<b>Фамилия:</b> %s\n", user.LastName)
    }

    if user.UserName != "" {
        info += fmt.Sprintf("<b>Username:</b> @%s\n", user.UserName)
    }

    info += fmt.Sprintf("<b>Язык:</b> %s\n", user.LanguageCode)
    info += fmt.Sprintf("<b>Бот:</b> %v\n", user.IsBot)

    reply := tgbotapi.NewMessage(chatID, info)
    reply.ParseMode = tgbotapi.ModeHTML
    _, err := bot.Send(reply)
    return err
}
```

---

## 7. Создаём диспетчер обработчиков

**Что делаем:**  
Создаём структуру, которая управляет всеми обработчиками и направляет команды к нужному обработчику.

**Создаём файл `internal/handler/dispatcher.go`:**
```go
package handler

import (
    "fmt"
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Dispatcher управляет обработчиками команд
type Dispatcher struct {
    handlers map[string]Handler  // Карта: команда -> обработчик
}

// NewDispatcher создаёт новый диспетчер
func NewDispatcher() *Dispatcher {
    return &Dispatcher{
        handlers: make(map[string]Handler),
    }
}

// Register регистрирует обработчик
func (d *Dispatcher) Register(handler Handler) {
    command := handler.Command()
    d.handlers[command] = handler
    log.Printf("Зарегистрирован обработчик команды /%s", command)
}

// HandleCommand обрабатывает команду, направляя её к соответствующему обработчику
func (d *Dispatcher) HandleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    command := msg.Command()

    // Ищем обработчик для команды
    handler, exists := d.handlers[command]
    if !exists {
        // Обработчик не найден — отправляем сообщение о неизвестной команде
        return d.handleUnknownCommand(bot, msg)
    }

    // Вызываем обработчик
    err := handler.Handle(bot, msg)
    if err != nil {
        log.Printf("Ошибка обработки команды /%s: %v", command, err)
        return err
    }

    return nil
}

// handleUnknownCommand обрабатывает неизвестные команды
func (d *Dispatcher) handleUnknownCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID
    text := "Неизвестная команда. Используйте /help для списка доступных команд."

    reply := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(reply)
    return err
}
```
```

**Разбор:**

- `map[string]Handler` — карта, которая хранит обработчики по имени команды.

- `make(map[string]Handler)` — создаём пустую карту.

- `Register(handler Handler)` — регистрирует обработчик в диспетчере.

- `exists` — флаг, который показывает, есть ли обработчик для команды.

- `handler.Handle(bot, msg)` — вызываем метод обработчика. Интерфейс `Handler` гарантирует, что у всех обработчиков есть этот метод.

---

## 8. Создаём обработчик обычных сообщений

**Что делаем:**  
Создаём обработчик для текстовых сообщений (не команд).

**Создаём файл `internal/handler/message_handler.go`:**
```go
package handler

import (
    "fmt"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler обрабатывает обычные текстовые сообщения
type MessageHandler struct{}

// NewMessageHandler создаёт новый обработчик сообщений
func NewMessageHandler() *MessageHandler {
    return &MessageHandler{}
}

// Handle обрабатывает текстовое сообщение
func (h *MessageHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID
    text := msg.Text

    // Простой эхо-ответ
    replyText := fmt.Sprintf("Вы написали: %s", text)
    
    reply := tgbotapi.NewMessage(chatID, replyText)
    _, err := bot.Send(reply)
    return err
}
```

---

## 9. Обновляем main.go

**Что делаем:**  
Обновляем главный файл, чтобы использовать новую модульную структуру.

**Обновляем файл `cmd/bot/main.go`:**
```go
package main

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/config"
    "telegram-bot/internal/handler"
)

func main() {
    // Загружаем конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Ошибка загрузки конфигурации:", err)
    }

    // Создаём экземпляр бота
    bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
    if err != nil {
        log.Fatal("Ошибка создания бота:", err)
    }

    bot.Debug = cfg.Bot.Debug
    log.Printf("Авторизован как %s", bot.Self.UserName)

    // Создаём диспетчер обработчиков
    dispatcher := handler.NewDispatcher()

    // Регистрируем обработчики команд
    dispatcher.Register(handler.NewStartHandler())
    dispatcher.Register(handler.NewHelpHandler())
    dispatcher.Register(handler.NewInfoHandler())

    // Создаём обработчик обычных сообщений
    messageHandler := handler.NewMessageHandler()

    // Настраиваем получение обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = cfg.Bot.Timeout
    updates := bot.GetUpdatesChan(u)

    // Обрабатываем обновления
    for update := range updates {
        handleUpdate(bot, dispatcher, messageHandler, update)
    }
}

// handleUpdate обрабатывает одно обновление
func handleUpdate(
    bot *tgbotapi.BotAPI,
    dispatcher *handler.Dispatcher,
    messageHandler *handler.MessageHandler,
    update tgbotapi.Update,
) {
    // Проверяем, есть ли сообщение
    if update.Message == nil {
        return
    }

    msg := update.Message

    // Обрабатываем команды
    if msg.IsCommand() {
        err := dispatcher.HandleCommand(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки команды: %v", err)
        }
        return
    }

    // Обрабатываем обычные сообщения
    if msg.Text != "" {
        err := messageHandler.Handle(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки сообщения: %v", err)
        }
    }
}
```

**Разбор:**

- Мы создаём диспетчер и регистрируем все обработчики перед началом цикла обработки обновлений.

- Функция `handleUpdate` теперь получает диспетчер и обработчик сообщений как параметры.

- Код стал более модульным — легко добавлять новые команды, просто создавая новый обработчик и регистрируя его.

---

## 10. Добавляем middleware для логирования

**Что такое middleware:**  
Middleware (промежуточное ПО) — это функции, которые выполняются до или после основной обработки. Например, логирование, проверка прав доступа.

**Создаём файл `internal/middleware/logger.go`:**
```go
package middleware

import (
    "log"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// LogCommand логирует команду перед обработкой
func LogCommand(msg *tgbotapi.Message) {
    user := msg.From
    command := msg.Command()
    
    log.Printf(
        "[%s] Команда /%s от пользователя %s (ID: %d) в чате %d",
        time.Now().Format("2006-01-02 15:04:05"),
        command,
        user.UserName,
        user.ID,
        msg.Chat.ID,
    )
}

// LogMessage логирует текстовое сообщение
func LogMessage(msg *tgbotapi.Message) {
    user := msg.From
    
    log.Printf(
        "[%s] Сообщение от пользователя %s (ID: %d): %s",
        time.Now().Format("2006-01-02 15:04:05"),
        user.UserName,
        user.ID,
        msg.Text,
    )
}
```

**Обновляем `handleUpdate` в `main.go`:**
```go
import (
    "telegram-bot/internal/middleware"
    // ... другие импорты
)

func handleUpdate(
    bot *tgbotapi.BotAPI,
    dispatcher *handler.Dispatcher,
    messageHandler *handler.MessageHandler,
    update tgbotapi.Update,
) {
    if update.Message == nil {
        return
    }

    msg := update.Message

    if msg.IsCommand() {
        middleware.LogCommand(msg)  // Логируем команду
        err := dispatcher.HandleCommand(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки команды: %v", err)
        }
        return
    }

    if msg.Text != "" {
        middleware.LogMessage(msg)  // Логируем сообщение
        err := messageHandler.Handle(bot, msg)
        if err != nil {
            log.Printf("Ошибка обработки сообщения: %v", err)
        }
    }
}
```

---

## 11. Добавляем middleware для проверки прав

**Что делаем:**  
Создаём middleware для проверки, является ли пользователь администратором.

**Создаём файл `internal/middleware/auth.go`:**
```go
package middleware

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsAdmin проверяет, является ли пользователь администратором
func IsAdmin(userID int64, adminIDs []int64) bool {
    for _, adminID := range adminIDs {
        if userID == adminID {
            return true
        }
    }
    return false
}

// RequireAdmin проверяет права доступа и отправляет сообщение, если пользователь не админ
func RequireAdmin(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, adminIDs []int64) bool {
    userID := msg.From.ID
    
    if !IsAdmin(userID, adminIDs) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды.")
        bot.Send(reply)
        return false
    }
    
    return true
}
```

**Пример использования в обработчике:**
```go
// AdminHandler обрабатывает админ-команду
type AdminHandler struct {
    adminIDs []int64
}

func (h *AdminHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    // Проверяем права доступа
    if !middleware.RequireAdmin(bot, msg, h.adminIDs) {
        return nil  // Сообщение уже отправлено middleware
    }
    
    // Продолжаем обработку команды
    reply := tgbotapi.NewMessage(msg.Chat.ID, "Админ-команда выполнена!")
    _, err := bot.Send(reply)
    return err
}
```

---

## 12. Структура проекта

**Как должна выглядеть структура проекта сейчас:**

```
telegram-bot/
├── cmd/
│   └── bot/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   └── bot.go
│   ├── handler/
│   │   ├── handler.go
│   │   ├── dispatcher.go
│   │   ├── start_handler.go
│   │   ├── help_handler.go
│   │   ├── info_handler.go
│   │   └── message_handler.go
│   └── middleware/
│       ├── logger.go
│       └── auth.go
├── .env
├── .gitignore
├── go.mod
└── go.sum
```

---

## Типичные ошибки

**Ошибка 1: "handler.Handle undefined"**

**Причина:** Структура не реализует интерфейс `Handler` (метод `Handle` или `Command` отсутствует).

**Решение:** Убедитесь, что у обработчика есть оба метода: `Handle` и `Command`.

**Ошибка 2: "cannot use handler (type *StartHandler) as Handler"**

**Причина:** Методы обработчика имеют неправильную сигнатуру (не совпадают с интерфейсом).

**Решение:** Проверьте, что методы имеют точно такие же параметры и возвращаемые значения, как в интерфейсе.

**Ошибка 3: Команда не обрабатывается**

**Причина:** Обработчик не зарегистрирован в диспетчере.

**Решение:** Убедитесь, что вы вызываете `dispatcher.Register()` для каждого обработчика в `main.go`.

---

## Что мы узнали

- Как создать модульную структуру обработчиков
- Что такое интерфейсы в Go и зачем они нужны
- Как создать диспетчер для управления обработчиками
- Как разделить код на отдельные файлы
- Как использовать middleware для логирования
- Как добавить проверку прав доступа
- Как сделать код более поддерживаемым и расширяемым

---

[Следующая глава: Работа с клавиатурами и инлайн-кнопками](./05-keyboards.md)

[Вернуться к оглавлению](./README.md)

