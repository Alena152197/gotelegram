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
