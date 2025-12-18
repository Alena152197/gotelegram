# Глава 9: Тестирование и отладка

В этой главе мы научимся тестировать бота и отлаживать возникающие проблемы. Правильное тестирование поможет избежать ошибок в продакшене.

---

## 1. Зачем нужно тестирование

**Преимущества тестирования:**
- Уверенность в работоспособности кода
- Быстрое обнаружение ошибок
- Упрощение рефакторинга
- Документация поведения кода

**Типы тестов:**
- Unit-тесты — тестирование отдельных функций
- Integration-тесты — тестирование взаимодействия компонентов

---

## 2. Unit-тесты для обработчиков

**Создаём файл `internal/handler/start_handler_test.go`:**
```go
package handler

import (
    "testing"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestStartHandler_Command проверяет, что обработчик возвращает правильную команду
func TestStartHandler_Command(t *testing.T) {
    handler := NewStartHandler()
    
    expected := "start"
    actual := handler.Command()
    
    if actual != expected {
        t.Errorf("Ожидалось %s, получено %s", expected, actual)
    }
}

// TestStartHandler_Handle проверяет обработку команды
func TestStartHandler_Handle(t *testing.T) {
    // Создаём mock бота и сообщения
    // В реальном проекте используйте библиотеку для моков, например testify
    
    handler := NewStartHandler()
    msg := &tgbotapi.Message{
        Chat: &tgbotapi.Chat{ID: 123456789},
    }
    
    // Здесь нужно использовать mock бота
    // err := handler.Handle(mockBot, msg)
    // if err != nil {
    //     t.Errorf("Ошибка обработки команды: %v", err)
    // }
}
```

**Запуск тестов:**
```bash
go test ./internal/handler/...
go test -v ./internal/handler/...  # С подробным выводом
go test -cover ./internal/handler/...  # С покрытием кода
```

---

## 3. Тестирование репозиториев

**Создаём тестовую базу данных:**
```go
// setupTestDB создаёт тестовую БД
func setupTestDB(t *testing.T) *sql.DB {
    // Подключаемся к тестовой БД
    db, err := sql.Open("postgres", "postgres://test:test@localhost/test_db?sslmode=disable")
    if err != nil {
        t.Fatal("Ошибка подключения к тестовой БД:", err)
    }
    
    // Выполняем миграции
    // ...
    
    return db
}

// TestUserRepository_CreateOrUpdate тестирует создание пользователя
func TestUserRepository_CreateOrUpdate(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    user := &domain.User{
        ID:        123456789,
        FirstName: "Test",
        Username:  "test_user",
    }
    
    err := repo.CreateOrUpdate(user)
    if err != nil {
        t.Errorf("Ошибка создания пользователя: %v", err)
    }
    
    // Проверяем, что пользователь создан
    savedUser, err := repo.GetByID(user.ID)
    if err != nil {
        t.Errorf("Ошибка получения пользователя: %v", err)
    }
    
    if savedUser.FirstName != user.FirstName {
        t.Errorf("Имя не совпадает: ожидалось %s, получено %s", 
            user.FirstName, savedUser.FirstName)
    }
}
```

---

## 4. Отладка бота

### Логирование

**Используем структурированное логирование:**
```go
import "go.uber.org/zap"

// В main.go
logger, err := zap.NewProduction()
if err != nil {
    log.Fatal("Ошибка создания логгера:", err)
}
defer logger.Sync()

// Логируем события
logger.Info("Бот запущен",
    zap.String("username", bot.Self.UserName),
    zap.String("token", cfg.Bot.Token[:10]+"..."), // Не логируем полный токен!
)
```

### Отладка в режиме разработки

**Включаем режим отладки:**
```go
bot.Debug = true  // Включает подробное логирование запросов к API
```

**Логируем обновления:**
```go
func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    // Логируем все обновления для отладки
    updateJSON, _ := json.MarshalIndent(update, "", "  ")
    log.Printf("Получено обновление:\n%s", updateJSON)
    
    // ... остальной код
}
```

---

## 5. Обработка ошибок

**Создаём централизованный обработчик ошибок:**
```go
// handleError обрабатывает ошибки
func handleError(logger *zap.Logger, err error, context string) {
    if err != nil {
        logger.Error("Ошибка",
            zap.String("контекст", context),
            zap.Error(err),
        )
        
        // Можно отправлять уведомления админам
        // notifyAdmins(err)
    }
}

// Использование:
err := dispatcher.HandleCommand(bot, msg)
if err != nil {
    handleError(logger, err, "обработка команды")
}
```

---

## 6. Типичные проблемы и решения

### Проблема 1: Бот не отвечает

**Диагностика:**
1. Проверьте, что бот запущен
2. Проверьте логи на наличие ошибок
3. Убедитесь, что токен правильный
4. Проверьте интернет-соединение

**Решение:**
```go
// Добавляем проверку соединения
func checkConnection(bot *tgbotapi.BotAPI) error {
    _, err := bot.GetMe()
    return err
}
```

### Проблема 2: Превышен лимит запросов

**Причина:** Слишком много запросов к API за короткое время.

**Решение:**
```go
import "time"

// Ограничиваем скорость отправки сообщений
var lastSentTime time.Time
var minDelay = time.Second / 30 // 30 сообщений в секунду (лимит Telegram)

func sendWithRateLimit(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
    elapsed := time.Since(lastSentTime)
    if elapsed < minDelay {
        time.Sleep(minDelay - elapsed)
    }
    
    _, err := bot.Send(msg)
    lastSentTime = time.Now()
    return err
}
```

### Проблема 3: Память утекает

**Причина:** Не освобождаются ресурсы (не закрываются соединения, не очищаются срезы).

**Решение:**
```go
// Всегда используйте defer для закрытия ресурсов
defer postgresDB.Close()
defer rows.Close()
defer file.Close()
```

---

## 7. Профилирование производительности

**Используем встроенный профайлер Go:**
```go
import (
    _ "net/http/pprof"
    "net/http"
)

// В main.go для профилирования
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Затем используйте:
// go tool pprof http://localhost:6060/debug/pprof/profile
```

---

## Что мы узнали

- Как писать unit-тесты для обработчиков
- Как тестировать репозитории
- Как использовать логирование для отладки
- Как обрабатывать ошибки
- Как диагностировать типичные проблемы
- Как профилировать производительность

---

[Следующая глава: Развёртывание бота](./10-deployment.md)

[Вернуться к оглавлению](./README.md)

