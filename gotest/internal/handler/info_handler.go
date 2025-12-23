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