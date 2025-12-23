package handler

import (
	"telegram-bot/internal/keyboard"

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

	// Прикрепляем клавиатуру к сообщению
	// reply.ReplyMarkup = keyboard.NewLanguageKeyboard()
    // сложная клавиатура
	// reply.ReplyMarkup = keyboard.NewMainMenuKeyboard()

	reply.ReplyMarkup = keyboard.NewConfirmKeyboard("delete_profile")
	bot.Send(reply)

	_, err := bot.Send(reply)
	return err
}
