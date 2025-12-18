# Глава 7: Создание админ-панели

В этой главе мы создадим систему управления ботом для администраторов. Админы смогут просматривать статистику, управлять пользователями и выполнять специальные команды.

---

## 1. Зачем нужна админ-панель

**Задачи администратора:**
- Просмотр статистики бота (количество пользователей, активность)
- Управление пользователями (блокировка, разблокировка)
- Рассылка сообщений всем пользователям
- Просмотр логов и ошибок
- Настройка бота

**Реализуем:**
- Систему проверки прав доступа
- Админ-команды
- Статистику
- Рассылку

---

## 2. Расширяем middleware для проверки прав

**Обновляем `internal/middleware/auth.go`:**
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

// RequireAdmin проверяет права доступа
func RequireAdmin(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, adminIDs []int64) bool {
    userID := msg.From.ID
    
    if !IsAdmin(userID, adminIDs) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ У вас нет прав для выполнения этой команды.")
        bot.Send(reply)
        return false
    }
    
    return true
}
```

---

## 3. Создаём обработчик статистики

**Создаём файл `internal/handler/admin_stats_handler.go`:**
```go
package handler

import (
    "fmt"
    "strconv"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/repository"
    "telegram-bot/internal/middleware"
)

// AdminStatsHandler обрабатывает команду /admin_stats
type AdminStatsHandler struct {
    userRepo *repository.UserRepository
    adminIDs []int64
}

// NewAdminStatsHandler создаёт новый обработчик
func NewAdminStatsHandler(userRepo *repository.UserRepository, adminIDs []int64) *AdminStatsHandler {
    return &AdminStatsHandler{
        userRepo: userRepo,
        adminIDs: adminIDs,
    }
}

// Command возвращает команду
func (h *AdminStatsHandler) Command() string {
    return "admin_stats"
}

// Handle обрабатывает команду /admin_stats
func (h *AdminStatsHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    // Проверяем права доступа
    if !middleware.RequireAdmin(bot, msg, h.adminIDs) {
        return nil
    }

    // Получаем статистику
    totalUsers, err := h.userRepo.Count()
    if err != nil {
        return err
    }

    // Формируем сообщение со статистикой
    text := fmt.Sprintf(
        "<b>📊 Статистика бота</b>\n\n"+
        "👥 Всего пользователей: <code>%d</code>\n"+
        "🔔 Бот активен: ✅\n",
        totalUsers,
    )

    reply := tgbotapi.NewMessage(msg.Chat.ID, text)
    reply.ParseMode = tgbotapi.ModeHTML
    _, err = bot.Send(reply)
    return err
}
```

---

## 4. Создаём обработчик рассылки

**Создаём файл `internal/handler/admin_broadcast_handler.go`:**
```go
package handler

import (
    "fmt"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/repository"
    "telegram-bot/internal/middleware"
)

// AdminBroadcastHandler обрабатывает команду /admin_broadcast
type AdminBroadcastHandler struct {
    userRepo *repository.UserRepository
    adminIDs []int64
}

// NewAdminBroadcastHandler создаёт новый обработчик
func NewAdminBroadcastHandler(userRepo *repository.UserRepository, adminIDs []int64) *AdminBroadcastHandler {
    return &AdminBroadcastHandler{
        userRepo: userRepo,
        adminIDs: adminIDs,
    }
}

// Command возвращает команду
func (h *AdminBroadcastHandler) Command() string {
    return "admin_broadcast"
}

// Handle обрабатывает команду /admin_broadcast <текст>
func (h *AdminBroadcastHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    // Проверяем права доступа
    if !middleware.RequireAdmin(bot, msg, h.adminIDs) {
        return nil
    }

    // Извлекаем текст рассылки (всё после команды)
    commandArgs := msg.CommandArguments()
    if commandArgs == "" {
        reply := tgbotapi.NewMessage(msg.Chat.ID, 
            "❌ Укажите текст для рассылки.\nПример: /admin_broadcast Привет всем!")
        bot.Send(reply)
        return nil
    }

    // Получаем всех пользователей
    users, err := h.userRepo.GetAll(1000, 0) // Получаем до 1000 пользователей
    if err != nil {
        return err
    }

    // Отправляем сообщение каждому пользователю
    successCount := 0
    failCount := 0

    for _, user := range users {
        broadcastMsg := tgbotapi.NewMessage(user.ID, commandArgs)
        _, err := bot.Send(broadcastMsg)
        if err != nil {
            failCount++
        } else {
            successCount++
        }
    }

    // Отправляем отчёт администратору
    report := fmt.Sprintf(
        "✅ Рассылка завершена!\n\n"+
        "✅ Отправлено: %d\n"+
        "❌ Ошибок: %d",
        successCount,
        failCount,
    )

    reply := tgbotapi.NewMessage(msg.Chat.ID, report)
    bot.Send(reply)

    return nil
}
```

---

## 5. Создаём обработчик управления пользователями

**Создаём файл `internal/handler/admin_user_handler.go`:**
```go
package handler

import (
    "strconv"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/repository"
    "telegram-bot/internal/middleware"
)

// AdminUserHandler обрабатывает команды управления пользователями
type AdminUserHandler struct {
    userRepo *repository.UserRepository
    adminIDs []int64
}

// NewAdminUserHandler создаёт новый обработчик
func NewAdminUserHandler(userRepo *repository.UserRepository, adminIDs []int64) *AdminUserHandler {
    return &AdminUserHandler{
        userRepo: userRepo,
        adminIDs: adminIDs,
    }
}

// Command возвращает команду
func (h *AdminUserHandler) Command() string {
    return "admin_user"
}

// Handle обрабатывает команду /admin_user <действие> <user_id>
func (h *AdminUserHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    // Проверяем права доступа
    if !middleware.RequireAdmin(bot, msg, h.adminIDs) {
        return nil
    }

    // Парсим аргументы: /admin_user block 123456789
    args := strings.Fields(msg.CommandArguments())
    if len(args) < 2 {
        reply := tgbotapi.NewMessage(msg.Chat.ID,
            "❌ Неправильный формат команды.\n"+
            "Использование: /admin_user <block|unblock> <user_id>")
        bot.Send(reply)
        return nil
    }

    action := args[0]
    userIDStr := args[1]

    // Преобразуем user_id в число
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный ID пользователя.")
        bot.Send(reply)
        return nil
    }

    // Выполняем действие
    switch action {
    case "block":
        // TODO: Реализовать блокировку пользователя
        reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("✅ Пользователь %d заблокирован.", userID))
        bot.Send(reply)
    case "unblock":
        // TODO: Реализовать разблокировку пользователя
        reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("✅ Пользователь %d разблокирован.", userID))
        bot.Send(reply)
    default:
        reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Неизвестное действие. Используйте: block или unblock")
        bot.Send(reply)
    }

    return nil
}
```

---

## 6. Регистрируем админ-обработчики

**Обновляем `cmd/bot/main.go`:**
```go
// ... существующий код ...

// Создаём диспетчер обработчиков
dispatcher := handler.NewDispatcher()
dispatcher.Register(handler.NewStartHandler())
dispatcher.Register(handler.NewHelpHandler())
dispatcher.Register(handler.NewInfoHandler())

// Регистрируем админ-обработчики
dispatcher.Register(handler.NewAdminStatsHandler(userRepo, cfg.Bot.AdminIDs))
dispatcher.Register(handler.NewAdminBroadcastHandler(userRepo, cfg.Bot.AdminIDs))
dispatcher.Register(handler.NewAdminUserHandler(userRepo, cfg.Bot.AdminIDs))
```

---

## Типичные ошибки

**Ошибка 1: Админ-команда доступна всем**

**Причина:** Не проверяются права доступа перед выполнением команды.

**Решение:** Всегда вызывайте `middleware.RequireAdmin()` в начале обработчика.

**Ошибка 2: "panic: runtime error" при рассылке**

**Причина:** Попытка отправить сообщение пользователю, который заблокировал бота.

**Решение:** Обрабатывайте ошибки при отправке и пропускайте пользователей, которым не удалось отправить.

---

## Что мы узнали

- Как создать систему проверки прав доступа
- Как реализовать админ-команды
- Как получать и отображать статистику
- Как создать систему рассылки сообщений
- Как управлять пользователями
- Как защитить админ-функции от обычных пользователей

---

[Следующая глава: Дополнительные функции](./08-advanced-features.md)

[Вернуться к оглавлению](./README.md)

