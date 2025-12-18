# Глава 3: Работа с Telegram Bot API

В этой главе мы создадим первого работающего бота, который может получать и отправлять сообщения. Это ключевая глава, которая даст основу для всего дальнейшего функционала.

---

## 1. Что такое Telegram Bot API

**Простое объяснение:**  
Telegram Bot API — это интерфейс для создания ботов в Telegram. Через этот API ваш код может отправлять сообщения, получать обновления (новые сообщения), работать с клавиатурами и многое другое.

**Как это работает:**
1. Вы создаёте бота через @BotFather в Telegram
2. Получаете токен (уникальный ключ для доступа к API)
3. Ваша программа отправляет HTTP-запросы к серверам Telegram с этим токеном
4. Telegram возвращает ответы в формате JSON

**Два режима работы:**
- **Long polling** (долгий опрос) — ваш бот постоянно спрашивает у Telegram: "Есть ли новые сообщения?"
- **Webhooks** (вебхуки) — Telegram сам отправляет обновления на ваш сервер

В этом руководстве мы будем использовать **long polling**, так как это проще для начала.

---

## 2. Создание бота через BotFather

**Что делаем:**  
Получаем токен бота от официального бота Telegram.

**Шаги:**
1. Откройте Telegram и найдите бота @BotFather
2. Отправьте команду `/newbot`
3. Введите имя бота (например, "My Test Bot")
4. Введите username бота (должен заканчиваться на `bot`, например `my_test_bot`)
5. BotFather выдаст вам токен в формате `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`

**Важно:**  
Никогда не публикуйте токен! Тот, кто знает токен, может управлять вашим ботом.

**Пример ответа BotFather:**
```
Done! Congratulations on your new bot. You will find it at t.me/my_test_bot. You can now add a description, about section and profile picture for your bot, see /help for a list of commands.

Use this token to access the HTTP API:
123456789:ABCdefGHIjklMNOpqrsTUVwxyz

Keep your token secure and store it safely, it can be used by anyone to control your bot.
```

**Токен выглядит так:** `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`

---

## 3. Устанавливаем библиотеку go-telegram-bot-api

**Что делаем:**  
Устанавливаем библиотеку для работы с Telegram Bot API.

**Команда:**
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

**Что такое go-telegram-bot-api:**  
Это популярная библиотека для Go, которая упрощает работу с Telegram Bot API. Она скрывает детали HTTP-запросов и предоставляет удобные функции.

---

## 4. Создаём первого бота

**Что делаем:**  
Создаём простого бота, который отправляет сообщение при получении команды `/start`.

**Создаём файл `cmd/bot/main.go`:**
```go
package main

import (
    "log"
    "os"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/config"
)

func main() {
    // Загружаем конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Ошибка загрузки конфигурации:", err)
    }

    // Создаём экземпляр бота с токеном
    bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
    if err != nil {
        log.Fatal("Ошибка создания бота:", err)
    }

    // Включаем режим отладки (если нужно)
    bot.Debug = cfg.Bot.Debug

    // Выводим информацию о боте
    log.Printf("Авторизован как %s", bot.Self.UserName)

    // Создаём канал для получения обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = cfg.Bot.Timeout

    // Получаем канал обновлений
    updates := bot.GetUpdatesChan(u)

    // Обрабатываем каждое обновление
    for update := range updates {
        // Обрабатываем обновление
        handleUpdate(bot, update)
    }
}
```

**Разбор:**

- `tgbotapi.NewBotAPI(cfg.Bot.Token)` — создаёт новый экземпляр бота с токеном. Эта функция отправляет запрос к Telegram API для проверки токена и получения информации о боте.

- `bot.Self.UserName` — имя пользователя бота (username, например, `my_test_bot`).

- `tgbotapi.NewUpdate(0)` — создаёт настройки для получения обновлений
  - `0` означает "получить все обновления с начала" (для продакшена лучше сохранять последний ID)

- `bot.GetUpdatesChan(u)` — получает канал (channel) обновлений. Это специальный тип в Go, который позволяет получать данные асинхронно.

- `for update := range updates` — перебираем все обновления из канала. Каждый раз, когда приходит новое сообщение, оно попадает в канал, и цикл продолжается.

**Что такое канал (channel) в Go:**  
Канал — это способ безопасно передавать данные между горутинами (параллельными потоками выполнения). В данном случае библиотека создаёт канал и постоянно заполняет его новыми обновлениями от Telegram.

---

## 5. Обработка обновлений

**Что делаем:**  
Создаём функцию для обработки входящих обновлений (сообщений, команд).

**Добавляем функцию `handleUpdate` в `cmd/bot/main.go`:**
```go
// handleUpdate обрабатывает одно обновление от Telegram
func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    // Проверяем, есть ли сообщение
    if update.Message == nil {
        return  // Если сообщения нет — игнорируем обновление
    }

    // Получаем сообщение и чат
    msg := update.Message
    chatID := msg.Chat.ID

    // Проверяем, является ли сообщение командой
    if msg.IsCommand() {
        handleCommand(bot, msg)
        return
    }

    // Если это не команда — отвечаем эхом
    reply := tgbotapi.NewMessage(chatID, "Вы написали: "+msg.Text)
    bot.Send(reply)
}
```

**Разбор:**

- `update.Message` — входящее сообщение. Может быть `nil`, если обновление не содержит сообщения (например, это callback-запрос от кнопки).

- `msg.Chat.ID` — уникальный идентификатор чата. Используется для отправки ответа в правильный чат.

- `msg.IsCommand()` — проверяет, является ли сообщение командой (начинается с `/`).

- `tgbotapi.NewMessage(chatID, text)` — создаёт новое сообщение для отправки.

- `bot.Send(reply)` — отправляет сообщение в Telegram.

---

## 6. Обработка команд

**Что делаем:**  
Создаём функцию для обработки команд бота (например, `/start`, `/help`).

**Добавляем функцию `handleCommand` в `cmd/bot/main.go`:**
```go
// handleCommand обрабатывает команды бота
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    chatID := msg.Chat.ID
    command := msg.Command()

    // Обрабатываем разные команды
    switch command {
    case "start":
        handleStartCommand(bot, chatID)
    case "help":
        handleHelpCommand(bot, chatID)
    default:
        handleUnknownCommand(bot, chatID)
    }
}

// handleStartCommand обрабатывает команду /start
func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Привет! Я тестовый бот на Go.\n\n" +
        "Доступные команды:\n" +
        "/start - начать работу\n" +
        "/help - помощь"

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}

// handleHelpCommand обрабатывает команду /help
func handleHelpCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Это справочная информация.\n\n" +
        "Бот создан с помощью библиотеки go-telegram-bot-api.\n" +
        "Исходный код: https://github.com/go-telegram-bot-api/telegram-bot-api"

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}

// handleUnknownCommand обрабатывает неизвестные команды
func handleUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Неизвестная команда. Используйте /help для списка команд."

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}
```

**Разбор:**

- `msg.Command()` — извлекает команду из сообщения. Например, из `/start` получится `"start"`.

- `switch command` — конструкция для множественного выбора. Похожа на `if-else`, но более читаемая для сравнения одной переменной со многими значениями.

- `case "start"` — если команда равна "start", выполняем код до следующего `case` или `default`.

- `default` — выполняется, если ни один `case` не подошёл.

**Почему отдельные функции:**  
Каждая команда обрабатывается в отдельной функции. Это делает код понятнее и проще поддерживать.

---

## 7. Полный код файла main.go

**Создаём файл `cmd/bot/main.go` полностью:**
```go
package main

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/config"
)

func main() {
    // Загружаем конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Ошибка загрузки конфигурации:", err)
    }

    // Создаём экземпляр бота с токеном
    bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
    if err != nil {
        log.Fatal("Ошибка создания бота:", err)
    }

    // Включаем режим отладки
    bot.Debug = cfg.Bot.Debug

    // Выводим информацию о боте
    log.Printf("Авторизован как %s", bot.Self.UserName)

    // Создаём канал для получения обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = cfg.Bot.Timeout

    // Получаем канал обновлений
    updates := bot.GetUpdatesChan(u)

    // Обрабатываем каждое обновление
    for update := range updates {
        handleUpdate(bot, update)
    }
}

// handleUpdate обрабатывает одно обновление от Telegram
func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    // Проверяем, есть ли сообщение
    if update.Message == nil {
        return
    }

    // Получаем сообщение и чат
    msg := update.Message
    chatID := msg.Chat.ID

    // Проверяем, является ли сообщение командой
    if msg.IsCommand() {
        handleCommand(bot, msg)
        return
    }

    // Если это не команда — отвечаем эхом
    reply := tgbotapi.NewMessage(chatID, "Вы написали: "+msg.Text)
    bot.Send(reply)
}

// handleCommand обрабатывает команды бота
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    chatID := msg.Chat.ID
    command := msg.Command()

    switch command {
    case "start":
        handleStartCommand(bot, chatID)
    case "help":
        handleHelpCommand(bot, chatID)
    default:
        handleUnknownCommand(bot, chatID)
    }
}

// handleStartCommand обрабатывает команду /start
func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Привет! Я тестовый бот на Go.\n\n" +
        "Доступные команды:\n" +
        "/start - начать работу\n" +
        "/help - помощь"

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}

// handleHelpCommand обрабатывает команду /help
func handleHelpCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Это справочная информация.\n\n" +
        "Бот создан с помощью библиотеки go-telegram-bot-api.\n" +
        "Исходный код: https://github.com/go-telegram-bot-api/telegram-bot-api"

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}

// handleUnknownCommand обрабатывает неизвестные команды
func handleUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) {
    text := "Неизвестная команда. Используйте /help для списка команд."

    msg := tgbotapi.NewMessage(chatID, text)
    bot.Send(msg)
}
```

---

## 8. Запускаем бота

**Шаги:**
1. Убедитесь, что в файле `.env` указан правильный токен:
   ```
   BOT_TOKEN=ваш_токен_здесь
   ```

2. Запустите бота:
   ```bash
   go run cmd/bot/main.go
   ```

3. Вы должны увидеть:
   ```
   2024/01/15 10:30:45 Авторизован как my_test_bot
   ```

4. Откройте Telegram, найдите вашего бота по username и отправьте команду `/start`

5. Бот должен ответить приветственным сообщением!

---

## 9. Отправка разных типов сообщений

**Что делаем:**  
Изучаем, как отправлять сообщения с разметкой, фото, документы и другие типы контента.

**Добавляем функцию для демонстрации разных типов сообщений:**
```go
// Примеры отправки разных типов сообщений
func sendMessageExamples(bot *tgbotapi.BotAPI, chatID int64) {
    // 1. Обычное текстовое сообщение
    msg1 := tgbotapi.NewMessage(chatID, "Это обычное сообщение")
    bot.Send(msg1)

    // 2. Сообщение с Markdown-разметкой
    msg2 := tgbotapi.NewMessage(chatID, "*Жирный текст*\n_Курсив_\n`Моноширинный`")
    msg2.ParseMode = tgbotapi.ModeMarkdown
    bot.Send(msg2)

    // 3. Сообщение с HTML-разметкой
    msg3 := tgbotapi.NewMessage(chatID, "<b>Жирный</b>\n<i>Курсив</i>\n<code>Код</code>")
    msg3.ParseMode = tgbotapi.ModeHTML
    bot.Send(msg3)

    // 4. Сообщение без предварительного просмотра ссылок
    msg4 := tgbotapi.NewMessage(chatID, "Ссылка: https://example.com")
    msg4.DisableWebPagePreview = true
    bot.Send(msg4)

    // 5. Ответ на конкретное сообщение
    msg5 := tgbotapi.NewMessage(chatID, "Это ответ на ваше сообщение")
    msg5.ReplyToMessageID = 123  // ID сообщения, на которое отвечаем
    bot.Send(msg5)
}
```

**Разбор:**

- `ParseMode` — режим парсинга разметки. Поддерживаются:
  - `tgbotapi.ModeMarkdown` — Markdown-разметка
  - `tgbotapi.ModeHTML` — HTML-разметка
  - `tgbotapi.ModeMarkdownV2` — новая версия Markdown

- `DisableWebPagePreview` — отключает предварительный просмотр ссылок (превью картинок, заголовков).

- `ReplyToMessageID` — ID сообщения, на которое мы отвечаем. Telegram покажет, что это ответ на конкретное сообщение.

---

## 10. Получение информации о пользователе

**Что делаем:**  
Учимся извлекать информацию о пользователе из сообщения.

**Добавляем команду `/info`:**
```go
// handleInfoCommand обрабатывает команду /info
func handleInfoCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    chatID := msg.Chat.ID
    user := msg.From

    // Формируем информацию о пользователе
    info := "Информация о вас:\n\n"
    info += fmt.Sprintf("ID: %d\n", user.ID)
    info += fmt.Sprintf("Имя: %s\n", user.FirstName)
    
    if user.LastName != "" {
        info += fmt.Sprintf("Фамилия: %s\n", user.LastName)
    }
    
    if user.UserName != "" {
        info += fmt.Sprintf("Username: @%s\n", user.UserName)
    }
    
    info += fmt.Sprintf("Язык: %s\n", user.LanguageCode)
    info += fmt.Sprintf("Бот: %v\n", user.IsBot)

    msg := tgbotapi.NewMessage(chatID, info)
    bot.Send(msg)
}
```

**Разбор:**

- `msg.From` — информация о пользователе, отправившем сообщение.

- Поля структуры `User`:
  - `ID` — уникальный идентификатор пользователя в Telegram (int64)
  - `FirstName` — имя пользователя
  - `LastName` — фамилия (может быть пустой)
  - `UserName` — username (может быть пустым)
  - `LanguageCode` — код языка (например, "ru", "en")
  - `IsBot` — является ли пользователь ботом

- `fmt.Sprintf` — форматирует строку с подстановкой значений (как `printf` в C).

**Важно:**  
User ID в Telegram — это уникальный идентификатор, который не меняется. Его можно использовать для хранения данных пользователя в базе данных.

---

## 11. Обработка ошибок при отправке

**Что делаем:**  
Добавляем обработку ошибок при отправке сообщений, чтобы бот не падал при сбоях.

**Создаём вспомогательную функцию:**
```go
// sendMessage безопасно отправляет сообщение с обработкой ошибок
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    
    _, err := bot.Send(msg)
    if err != nil {
        log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
    }
}
```

**Разбор:**

- `bot.Send(msg)` возвращает два значения: отправленное сообщение и ошибку.

- Мы используем `_` (пустой идентификатор) для первого значения, так как нам не нужно отправленное сообщение.

- Если ошибка произошла, мы логируем её, но не падаем. Это позволяет боту продолжать работать.

**Когда могут возникнуть ошибки:**
- Пользователь заблокировал бота
- Проблемы с сетью
- Неправильный формат сообщения
- Превышен лимит отправки сообщений

---

## 12. Полный код с улучшениями

**Обновлённый файл `cmd/bot/main.go` с обработкой ошибок:**
```go
package main

import (
    "fmt"
    "log"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "telegram-bot/internal/config"
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

    // Настраиваем получение обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = cfg.Bot.Timeout
    updates := bot.GetUpdatesChan(u)

    // Обрабатываем обновления
    for update := range updates {
        handleUpdate(bot, update)
    }
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    msg := update.Message
    chatID := msg.Chat.ID

    if msg.IsCommand() {
        handleCommand(bot, msg)
        return
    }

    // Эхо-ответ на обычное сообщение
    sendMessage(bot, chatID, "Вы написали: "+msg.Text)
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    command := msg.Command()

    switch command {
    case "start":
        handleStartCommand(bot, msg)
    case "help":
        handleHelpCommand(bot, msg)
    case "info":
        handleInfoCommand(bot, msg)
    default:
        handleUnknownCommand(bot, msg)
    }
}

func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    text := "Привет! Я тестовый бот на Go.\n\n" +
        "Доступные команды:\n" +
        "/start - начать работу\n" +
        "/help - помощь\n" +
        "/info - информация о вас"

    sendMessage(bot, msg.Chat.ID, text)
}

func handleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    text := "Это справочная информация.\n\n" +
        "Бот создан с помощью библиотеки go-telegram-bot-api.\n" +
        "Исходный код: https://github.com/go-telegram-bot-api/telegram-bot-api"

    sendMessage(bot, msg.Chat.ID, text)
}

func handleInfoCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    user := msg.From
    chatID := msg.Chat.ID

    info := "Информация о вас:\n\n"
    info += fmt.Sprintf("ID: %d\n", user.ID)
    info += fmt.Sprintf("Имя: %s\n", user.FirstName)

    if user.LastName != "" {
        info += fmt.Sprintf("Фамилия: %s\n", user.LastName)
    }

    if user.UserName != "" {
        info += fmt.Sprintf("Username: @%s\n", user.UserName)
    }

    info += fmt.Sprintf("Язык: %s\n", user.LanguageCode)
    info += fmt.Sprintf("Бот: %v\n", user.IsBot)

    sendMessage(bot, chatID, info)
}

func handleUnknownCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    text := "Неизвестная команда. Используйте /help для списка команд."
    sendMessage(bot, msg.Chat.ID, text)
}

// sendMessage безопасно отправляет сообщение
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(msg)
    if err != nil {
        log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
    }
}
```

---

## Типичные ошибки

**Ошибка 1: "Unauthorized" или "401 Unauthorized"**

**Причина:** Неправильный токен бота.

**Решение:**
1. Проверьте, что токен в `.env` файле правильный
2. Убедитесь, что нет лишних пробелов
3. Попробуйте создать нового бота через @BotFather

**Ошибка 2: "Too Many Requests" или "429"**

**Причина:** Превышен лимит запросов к API (слишком много сообщений за короткое время).

**Решение:**
1. Добавьте задержки между отправкой сообщений
2. Используйте `time.Sleep()` для ограничения скорости

**Ошибка 3: Бот не отвечает**

**Причина:** Возможные проблемы:
- Бот не запущен
- Проблемы с сетью
- Неправильная обработка обновлений

**Решение:**
1. Проверьте логи на наличие ошибок
2. Убедитесь, что бот запущен и видит сообщение в логах
3. Проверьте интернет-соединение

---

## Что мы узнали

- Как получить токен бота через @BotFather
- Как создать бота с помощью библиотеки go-telegram-bot-api
- Как получать и обрабатывать обновления
- Как отправлять сообщения пользователям
- Как обрабатывать команды
- Как извлекать информацию о пользователе
- Как обрабатывать ошибки при отправке сообщений
- Как использовать разметку в сообщениях

---

[Следующая глава: Обработка команд и сообщений](./04-handlers.md)

[Вернуться к оглавлению](./README.md)

