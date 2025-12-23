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