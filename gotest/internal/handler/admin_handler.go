package handler

import (
	"fmt"
	"telegram-bot/internal/middleware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InfoHandler обрабатывает команду /info
// AdminHandler обрабатывает админ-команду
type AdminHandler struct {
	adminIDs []int64
}

// NewInfoHandler создаёт новый обработчик команды /admin
func NewAdminHandler(adminIDs []int64) *AdminHandler {
	return &AdminHandler{
		adminIDs: adminIDs,
	}
}

// Command возвращает команду
func (h *AdminHandler) Command() string {
	return "admin"
}

// Handle обрабатывает команду /info
func (h *AdminHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	if !middleware.RequireAdmin(bot, msg, h.adminIDs) {
		return nil // Сообщение уже отправлено middleware
	}
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
