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
// Кнопки Профиль, Настройки, Меню, Скрыть
func NewMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// Первый ряд
	btnProfile := tgbotapi.NewKeyboardButton("👤 Профиль")
	btnSettings := tgbotapi.NewKeyboardButton("⚙️ Настройки")
	row1 := tgbotapi.NewKeyboardButtonRow(btnProfile, btnSettings)

	// Второй ряд
	btnMenu := tgbotapi.NewKeyboardButton("📋 Меню")
	btnHide := tgbotapi.NewKeyboardButton("🔽 Скрыть")
	row2 := tgbotapi.NewKeyboardButtonRow(btnMenu, btnHide)

	// Создаём клавиатуру из всех рядов
	keyboard := tgbotapi.NewReplyKeyboard(row1, row2)
	keyboard.ResizeKeyboard = true // Автоматически подстраиваем размер кнопок

	return keyboard
}
