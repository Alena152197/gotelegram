package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/keyboard"
)

// HelpHandler обрабатывает команду /help
type HelpHandler struct{}

// NewHelpHandler создаёт новый обработчик команды /help
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

// Command возвращает команду
func (h *HelpHandler) Command() string {
	return "help"
}

// Handle обрабатывает команду /help
func (h *HelpHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	chatID := msg.Chat.ID

	text := "Это справочная информация.\n\n" +
		"<b>Доступные команды:</b>\n\n" +
		"/start - начать работу с ботом\n" +
		"/info - информация о вашем профиле\n\n" +
		"<b>Важно:</b> Если вы нажали на кнопку \"🔽 Скрыть\" и клавиатура исчезла, " +
		"нажмите /start - начать работу с ботом, и клавиатура снова появится."

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = tgbotapi.ModeHTML
	// Показываем клавиатуру, если она не скрыта
	reply.ReplyMarkup = keyboard.NewMainMenuKeyboard()
	_, err := bot.Send(reply)
	return err
}
