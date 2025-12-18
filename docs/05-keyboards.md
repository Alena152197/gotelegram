# Глава 5: Работа с клавиатурами и инлайн-кнопками

В этой главе мы научимся создавать интерактивные клавиатуры для бота. Это сделает взаимодействие с ботом более удобным и интуитивным.

---

## 1. Типы клавиатур в Telegram

**Telegram поддерживает три типа клавиатур:**

1. **Reply Keyboard** (клавиатура ответов) — постоянные кнопки, которые отображаются внизу экрана вместо обычной клавиатуры
2. **Inline Keyboard** (инлайн-клавиатура) — кнопки под конкретным сообщением
3. **Keyboard Commands** (команды клавиатуры) — список команд, который появляется при вводе `/`

**В этой главе мы разберём первые два типа.**

---

## 2. Reply Keyboard (клавиатура ответов)

**Что это:**  
Клавиатура ответов — это кнопки, которые заменяют обычную клавиатуру телефона. Когда пользователь нажимает кнопку, бот получает текст кнопки как обычное сообщение.

**Когда использовать:**
- Для быстрого выбора из ограниченного набора вариантов
- Для навигации по меню бота
- Когда нужно упростить ввод для пользователя

**Недостатки:**
- Клавиатура занимает место на экране
- Пользователь может её скрыть вручную

---

## 3. Создаём простую клавиатуру

**Что делаем:**  
Создаём клавиатуру с кнопками для выбора языка.

**Создаём файл `internal/keyboard/reply_keyboard.go`:**
```go
package keyboard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// NewLanguageKeyboard создаёт клавиатуру для выбора языка
func NewLanguageKeyboard() tgbotapi.ReplyKeyboardMarkup {
    // Создаём кнопки
    btnRussian := tgbotapi.NewKeyboardButton("🇷🇺 Русский")
    btnEnglish := tgbotapi.NewKeyboardButton("🇬🇧 English")
    
    // Создаём ряд кнопок (все кнопки в одном ряду)
    row := tgbotapi.NewKeyboardButtonRow(btnRussian, btnEnglish)
    
    // Создаём клавиатуру из рядов
    keyboard := tgbotapi.NewReplyKeyboard(row)
    
    return keyboard
}
```

**Разбор:**

- `tgbotapi.NewKeyboardButton("текст")` — создаёт кнопку с указанным текстом.

- `tgbotapi.NewKeyboardButtonRow(...)` — создаёт ряд кнопок. Все переданные кнопки будут в одной строке.

- `tgbotapi.NewReplyKeyboard(...)` — создаёт клавиатуру из рядов кнопок.

**Использование:**
```go
// В обработчике команды
func (h *SettingsHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
    chatID := msg.Chat.ID
    
    text := "Выберите язык:"
    reply := tgbotapi.NewMessage(chatID, text)
    
    // Прикрепляем клавиатуру к сообщению
    reply.ReplyMarkup = keyboard.NewLanguageKeyboard()
    
    _, err := bot.Send(reply)
    return err
}
```

---

## 4. Создаём клавиатуру с несколькими рядами

**Что делаем:**  
Создаём более сложную клавиатуру с несколькими рядами кнопок.

**Добавляем функцию в `internal/keyboard/reply_keyboard.go`:**
```go
// NewMainMenuKeyboard создаёт главное меню бота
func NewMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
    // Первый ряд
    btnProfile := tgbotapi.NewKeyboardButton("👤 Профиль")
    btnSettings := tgbotapi.NewKeyboardButton("⚙️ Настройки")
    row1 := tgbotapi.NewKeyboardButtonRow(btnProfile, btnSettings)
    
    // Второй ряд
    btnHelp := tgbotapi.NewKeyboardButton("❓ Помощь")
    btnAbout := tgbotapi.NewKeyboardButton("ℹ️ О боте")
    row2 := tgbotapi.NewKeyboardButtonRow(btnHelp, btnAbout)
    
    // Третий ряд (одна кнопка на весь ряд)
    btnCancel := tgbotapi.NewKeyboardButton("❌ Отмена")
    row3 := tgbotapi.NewKeyboardButtonRow(btnCancel)
    
    // Создаём клавиатуру из всех рядов
    keyboard := tgbotapi.NewReplyKeyboard(row1, row2, row3)
    
    return keyboard
}
```

**Разбор:**

- Каждый ряд создаётся отдельно с помощью `NewKeyboardButtonRow`.

- Количество кнопок в ряду может быть разным — от одной до нескольких.

- Все ряды передаются в `NewReplyKeyboard` через запятую.

---

## 5. Скрытие клавиатуры

**Что делаем:**  
Учимся скрывать клавиатуру, когда она больше не нужна.

**Добавляем функцию:**
```go
// RemoveKeyboard удаляет клавиатуру
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
    return tgbotapi.NewReplyKeyboardRemove()
}
```

**Использование:**
```go
// Отправляем сообщение и скрываем клавиатуру
reply := tgbotapi.NewMessage(chatID, "Клавиатура скрыта")
reply.ReplyMarkup = keyboard.RemoveKeyboard()
bot.Send(reply)
```

**Вариант с параметрами:**
```go
// RemoveKeyboard с опцией "Selective" (скрыть только для определённого пользователя)
func RemoveKeyboardSelective() tgbotapi.ReplyKeyboardRemove {
    remove := tgbotapi.NewReplyKeyboardRemove()
    remove.Selective = true  // Скрыть только для пользователя, которому отвечаем
    return remove
}
```

---

## 6. Inline Keyboard (инлайн-клавиатура)

**Что это:**  
Инлайн-клавиатура — это кнопки, которые появляются под конкретным сообщением. Они не занимают место постоянно и удобны для интерактивных действий.

**Преимущества:**
- Не занимают место постоянно
- Можно обновлять кнопки, редактируя сообщение
- Можно отправлять callback-запросы без текстовых сообщений
- Поддерживают URL-кнопки

**Когда использовать:**
- Для подтверждения действий
- Для навигации по страницам
- Для выбора опций без отправки нового сообщения
- Для ссылок на веб-страницы

---

## 7. Создаём простую инлайн-клавиатуру

**Что делаем:**  
Создаём инлайн-клавиатуру с кнопками для подтверждения действия.

**Создаём файл `internal/keyboard/inline_keyboard.go`:**
```go
package keyboard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// NewConfirmKeyboard создаёт клавиатуру с кнопками "Да" и "Нет"
func NewConfirmKeyboard(dataPrefix string) tgbotapi.InlineKeyboardMarkup {
    // Создаём инлайн-кнопки
    btnYes := tgbotapi.NewInlineKeyboardButtonData("✅ Да", dataPrefix+"_yes")
    btnNo := tgbotapi.NewInlineKeyboardButtonData("❌ Нет", dataPrefix+"_no")
    
    // Создаём ряд кнопок
    row := tgbotapi.NewInlineKeyboardRow(btnYes, btnNo)
    
    // Создаём клавиатуру
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    
    return keyboard
}
```

**Разбор:**

- `tgbotapi.NewInlineKeyboardButtonData("текст", "данные")` — создаёт инлайн-кнопку
  - Первый параметр — текст на кнопке
  - Второй параметр — данные (callback_data), которые будут отправлены боту при нажатии

- `dataPrefix` — префикс для идентификации разных типов callback-запросов. Например, для удаления профиля можно использовать `"delete_profile"`, а данные будут `"delete_profile_yes"` и `"delete_profile_no"`.

**Использование:**
```go
reply := tgbotapi.NewMessage(chatID, "Вы уверены, что хотите удалить профиль?")
reply.ReplyMarkup = keyboard.NewConfirmKeyboard("delete_profile")
bot.Send(reply)
```

---

## 8. Обработка callback-запросов

**Что это:**  
Когда пользователь нажимает инлайн-кнопку, Telegram отправляет боту callback-запрос. Бот должен обработать его и ответить.

**Что делаем:**  
Добавляем обработку callback-запросов в главный цикл бота.

**Обновляем `handleUpdate` в `cmd/bot/main.go`:**
```go
func handleUpdate(
    bot *tgbotapi.BotAPI,
    dispatcher *handler.Dispatcher,
    messageHandler *handler.MessageHandler,
    update tgbotapi.Update,
) {
    // Обрабатываем callback-запросы (нажатия на инлайн-кнопки)
    if update.CallbackQuery != nil {
        handleCallbackQuery(bot, update.CallbackQuery)
        return
    }

    // Обрабатываем сообщения
    if update.Message == nil {
        return
    }

    msg := update.Message

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

// handleCallbackQuery обрабатывает нажатие на инлайн-кнопку
func handleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
    data := callback.Data
    chatID := callback.Message.Chat.ID
    messageID := callback.Message.MessageID
    userID := callback.From.ID

    log.Printf("Callback от пользователя %d: %s", userID, data)

    // Отвечаем на callback-запрос (обязательно!)
    // Это уберёт индикатор загрузки на кнопке
    callbackConfig := tgbotapi.NewCallback(callback.ID, "")
    bot.Request(callbackConfig)

    // Обрабатываем данные в зависимости от префикса
    if strings.HasPrefix(data, "delete_profile_") {
        if data == "delete_profile_yes" {
            // Пользователь подтвердил удаление
            editText := "✅ Профиль удалён!"
            edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
            bot.Send(edit)
        } else if data == "delete_profile_no" {
            // Пользователь отменил удаление
            editText := "❌ Удаление отменено."
            edit := tgbotapi.NewEditMessageText(chatID, messageID, editText)
            bot.Send(edit)
        }
    }
}
```

**Разбор:**

- `update.CallbackQuery` — структура с информацией о нажатии на кнопку.

- `callback.Data` — данные кнопки (то, что мы указали при создании).

- `bot.Request(callbackConfig)` — отправляем ответ на callback-запрос. Это обязательно — иначе Telegram будет показывать индикатор загрузки на кнопке.

- `tgbotapi.NewEditMessageText` — редактирует существующее сообщение вместо отправки нового.

- `strings.HasPrefix` — проверяет, начинается ли строка с определённого префикса.

---

## 9. Создаём клавиатуру с URL-кнопками

**Что делаем:**  
Создаём кнопки, которые открывают веб-страницы.

**Добавляем функцию:**
```go
// NewURLKeyboard создаёт клавиатуру с URL-кнопками
func NewURLKeyboard() tgbotapi.InlineKeyboardMarkup {
    btnWebsite := tgbotapi.NewInlineKeyboardButtonURL("🌐 Сайт", "https://example.com")
    btnGitHub := tgbotapi.NewInlineKeyboardButtonURL("💻 GitHub", "https://github.com")
    
    row := tgbotapi.NewInlineKeyboardRow(btnWebsite, btnGitHub)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    
    return keyboard
}
```

**Разбор:**

- `NewInlineKeyboardButtonURL` — создаёт кнопку, которая открывает URL в браузере.

- URL-кнопки не отправляют callback-запросы, они просто открывают ссылку.

---

## 10. Создаём клавиатуру с кнопкой "Переключить"

**Что делаем:**  
Создаём кнопку, которая переключает состояние (например, включить/выключить уведомления).

**Добавляем функцию:**
```go
// NewToggleKeyboard создаёт клавиатуру с кнопкой переключения
func NewToggleKeyboard(enabled bool, callbackData string) tgbotapi.InlineKeyboardMarkup {
    var btnText string
    if enabled {
        btnText = "🔔 Уведомления: Включено"
    } else {
        btnText = "🔕 Уведомления: Выключено"
    }
    
    btn := tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData)
    row := tgbotapi.NewInlineKeyboardRow(btn)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    
    return keyboard
}
```

**Использование:**
```go
// Создаём клавиатуру с текущим состоянием
enabled := true
keyboard := keyboard.NewToggleKeyboard(enabled, "toggle_notifications")

reply := tgbotapi.NewMessage(chatID, "Настройки уведомлений:")
reply.ReplyMarkup = keyboard
bot.Send(reply)
```

**Обработка переключения:**
```go
if data == "toggle_notifications" {
    // Определяем текущее состояние (можно хранить в базе данных)
    currentState := true
    
    // Переключаем состояние
    newState := !currentState
    
    // Обновляем клавиатуру
    newKeyboard := keyboard.NewToggleKeyboard(newState, "toggle_notifications")
    edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, newKeyboard)
    bot.Send(edit)
}
```

**Разбор:**

- `NewEditMessageReplyMarkup` — обновляет только клавиатуру сообщения, не меняя текст.

---

## 11. Создаём пагинацию (навигацию по страницам)

**Что делаем:**  
Создаём клавиатуру с кнопками "Назад" и "Вперёд" для навигации по страницам.

**Создаём функцию:**
```go
// NewPaginationKeyboard создаёт клавиатуру с пагинацией
func NewPaginationKeyboard(currentPage, totalPages int, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
    var buttons []tgbotapi.InlineKeyboardButton
    
    // Кнопка "Назад"
    if currentPage > 1 {
        prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", 
            fmt.Sprintf("%s_page_%d", callbackPrefix, currentPage-1))
        buttons = append(buttons, prevBtn)
    }
    
    // Информация о странице
    pageInfo := tgbotapi.NewInlineKeyboardButtonData(
        fmt.Sprintf("Страница %d из %d", currentPage, totalPages),
        "page_info", // Неактивная кнопка (можно использовать "callback_data", который не обрабатывается)
    )
    buttons = append(buttons, pageInfo)
    
    // Кнопка "Вперёд"
    if currentPage < totalPages {
        nextBtn := tgbotapi.NewInlineKeyboardButtonData("Вперёд ➡️", 
            fmt.Sprintf("%s_page_%d", callbackPrefix, currentPage+1))
        buttons = append(buttons, nextBtn)
    }
    
    row := tgbotapi.NewInlineKeyboardRow(buttons...)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    
    return keyboard
}
```

**Разбор:**

- `buttons...` — оператор spread в Go. Разворачивает срез кнопок как отдельные аргументы.

- Условная логика показывает кнопки только когда они имеют смысл (не показываем "Назад" на первой странице).

---

## 12. Полный код файлов клавиатур

**Файл `internal/keyboard/reply_keyboard.go`:**
```go
package keyboard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// NewLanguageKeyboard создаёт клавиатуру для выбора языка
func NewLanguageKeyboard() tgbotapi.ReplyKeyboardMarkup {
    btnRussian := tgbotapi.NewKeyboardButton("🇷🇺 Русский")
    btnEnglish := tgbotapi.NewKeyboardButton("🇬🇧 English")
    row := tgbotapi.NewKeyboardButtonRow(btnRussian, btnEnglish)
    keyboard := tgbotapi.NewReplyKeyboard(row)
    return keyboard
}

// NewMainMenuKeyboard создаёт главное меню бота
func NewMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
    btnProfile := tgbotapi.NewKeyboardButton("👤 Профиль")
    btnSettings := tgbotapi.NewKeyboardButton("⚙️ Настройки")
    row1 := tgbotapi.NewKeyboardButtonRow(btnProfile, btnSettings)
    
    btnHelp := tgbotapi.NewKeyboardButton("❓ Помощь")
    btnAbout := tgbotapi.NewKeyboardButton("ℹ️ О боте")
    row2 := tgbotapi.NewKeyboardButtonRow(btnHelp, btnAbout)
    
    btnCancel := tgbotapi.NewKeyboardButton("❌ Отмена")
    row3 := tgbotapi.NewKeyboardButtonRow(btnCancel)
    
    keyboard := tgbotapi.NewReplyKeyboard(row1, row2, row3)
    return keyboard
}

// RemoveKeyboard удаляет клавиатуру
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
    return tgbotapi.NewReplyKeyboardRemove()
}
```

**Файл `internal/keyboard/inline_keyboard.go`:**
```go
package keyboard

import (
    "fmt"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewConfirmKeyboard создаёт клавиатуру с кнопками "Да" и "Нет"
func NewConfirmKeyboard(dataPrefix string) tgbotapi.InlineKeyboardMarkup {
    btnYes := tgbotapi.NewInlineKeyboardButtonData("✅ Да", dataPrefix+"_yes")
    btnNo := tgbotapi.NewInlineKeyboardButtonData("❌ Нет", dataPrefix+"_no")
    row := tgbotapi.NewInlineKeyboardRow(btnYes, btnNo)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    return keyboard
}

// NewURLKeyboard создаёт клавиатуру с URL-кнопками
func NewURLKeyboard() tgbotapi.InlineKeyboardMarkup {
    btnWebsite := tgbotapi.NewInlineKeyboardButtonURL("🌐 Сайт", "https://example.com")
    btnGitHub := tgbotapi.NewInlineKeyboardButtonURL("💻 GitHub", "https://github.com")
    row := tgbotapi.NewInlineKeyboardRow(btnWebsite, btnGitHub)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    return keyboard
}

// NewToggleKeyboard создаёт клавиатуру с кнопкой переключения
func NewToggleKeyboard(enabled bool, callbackData string) tgbotapi.InlineKeyboardMarkup {
    var btnText string
    if enabled {
        btnText = "🔔 Уведомления: Включено"
    } else {
        btnText = "🔕 Уведомления: Выключено"
    }
    
    btn := tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData)
    row := tgbotapi.NewInlineKeyboardRow(btn)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    return keyboard
}

// NewPaginationKeyboard создаёт клавиатуру с пагинацией
func NewPaginationKeyboard(currentPage, totalPages int, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
    var buttons []tgbotapi.InlineKeyboardButton
    
    if currentPage > 1 {
        prevBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", 
            fmt.Sprintf("%s_page_%d", callbackPrefix, currentPage-1))
        buttons = append(buttons, prevBtn)
    }
    
    pageInfo := tgbotapi.NewInlineKeyboardButtonData(
        fmt.Sprintf("Страница %d из %d", currentPage, totalPages),
        "page_info",
    )
    buttons = append(buttons, pageInfo)
    
    if currentPage < totalPages {
        nextBtn := tgbotapi.NewInlineKeyboardButtonData("Вперёд ➡️", 
            fmt.Sprintf("%s_page_%d", callbackPrefix, currentPage+1))
        buttons = append(buttons, nextBtn)
    }
    
    row := tgbotapi.NewInlineKeyboardRow(buttons...)
    keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
    return keyboard
}
```

---

## Типичные ошибки

**Ошибка 1: "Bad Request: BUTTON_DATA_INVALID"**

**Причина:** Callback data слишком длинный (максимум 64 байта).

**Решение:** Используйте короткие идентификаторы, а детали храните в базе данных.

**Ошибка 2: Кнопки не отображаются**

**Причина:** Неправильная структура клавиатуры или слишком много кнопок в ряду.

**Решение:** 
- Убедитесь, что вы используете правильные функции создания кнопок
- Не больше 3-4 кнопок в ряду для reply-клавиатуры
- Не больше 8 кнопок в ряду для inline-клавиатуры

**Ошибка 3: Callback-запрос не обрабатывается**

**Причина:** Не обработан `update.CallbackQuery` в главном цикле.

**Решение:** Добавьте проверку `if update.CallbackQuery != nil` перед обработкой сообщений.

**Ошибка 4: Индикатор загрузки на кнопке не исчезает**

**Причина:** Не отправлен ответ на callback-запрос.

**Решение:** Всегда вызывайте `bot.Request(callbackConfig)` после обработки callback-запроса.

---

## Что мы узнали

- Как создавать reply-клавиатуры (постоянные кнопки)
- Как создавать inline-клавиатуры (кнопки под сообщением)
- Как обрабатывать нажатия на кнопки (callback-запросы)
- Как создавать URL-кнопки
- Как обновлять клавиатуры динамически
- Как создавать пагинацию
- Как скрывать клавиатуры
- Типичные ошибки и способы их решения

---

[Следующая глава: Работа с базой данных](./06-database.md)

[Вернуться к оглавлению](./README.md)

