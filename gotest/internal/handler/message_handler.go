package handler

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/keyboard"
)

// MessageHandler обрабатывает обычные текстовые сообщения
type MessageHandler struct{}

// NewMessageHandler создаёт новый обработчик сообщений
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

// Handle обрабатывает текстовое сообщение
// getNotificationState - функция для получения состояния уведомлений пользователя
// getLanguage - функция для получения языка пользователя
func (h *MessageHandler) Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, getNotificationState func(int64) bool, getLanguage func(int64) string) error {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	// Обрабатываем команды из reply-клавиатуры
	switch text {
	case "👤 Профиль":
		return h.handleProfile(bot, chatID)

	case "⚙️ Настройки":
		return h.handleSettings(bot, msg, getNotificationState, getLanguage)

	case "📋 Меню":
		return h.handleMenu(bot, chatID)

	case "🔽 Скрыть":
		return h.handleHideKeyboard(bot, chatID)

	default:
		// Обработка других текстовых сообщений
		if strings.Contains(strings.ToLower(text), "подпис") {
			reply := tgbotapi.NewMessage(chatID, "Напишите администратору @Alex152197 — он с радостью вам поможет! 😊")
			reply.ReplyMarkup = keyboard.NewMainMenuKeyboard()
			_, err := bot.Send(reply)
			return err
		}

		// Эхо-ответ для остальных сообщений
		replyText := fmt.Sprintf("Вы написали: %s\n\nИспользуйте меню для навигации.", text)
		reply := tgbotapi.NewMessage(chatID, replyText)
		reply.ReplyMarkup = keyboard.NewMainMenuKeyboard()
		_, err := bot.Send(reply)
		return err
	}
}

// handleProfile обрабатывает запрос профиля
func (h *MessageHandler) handleProfile(bot *tgbotapi.BotAPI, chatID int64) error {
	user := bot.Self
	text := fmt.Sprintf("👤 Ваш профиль:\n\n"+
		"ID: %d\n"+
		"Имя: %s\n"+
		"Username: @%s\n\n"+
		"Хотите удалить профиль?", chatID, user.FirstName, user.UserName)

	reply := tgbotapi.NewMessage(chatID, text)
	// Показываем inline-клавиатуру с кнопками подтверждения удаления
	kb := keyboard.NewConfirmKeyboard("delete_profile")
	kb = keyboard.AddBackButton(kb)
	reply.ReplyMarkup = &kb
	_, err := bot.Send(reply)
	return err
}

// handleSettings обрабатывает запрос настроек
func (h *MessageHandler) handleSettings(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, getNotificationState func(int64) bool, getLanguage func(int64) string) error {
	chatID := msg.Chat.ID

	text := "⚙️ Настройки:\n\n" +
		"Выберите настройку для изменения:"

	reply := tgbotapi.NewMessage(chatID, text)

	// Создаём inline-клавиатуру с кнопками настроек
	btnNotif := tgbotapi.NewInlineKeyboardButtonData("🔔 Уведомления", "settings_notif")
	btnLang := tgbotapi.NewInlineKeyboardButtonData("🌐 Язык", "settings_lang")
	row := tgbotapi.NewInlineKeyboardRow(btnNotif, btnLang)
	kb := tgbotapi.NewInlineKeyboardMarkup(row)
	kb = keyboard.AddBackButton(kb)
	reply.ReplyMarkup = &kb

	_, err := bot.Send(reply)
	return err
}

// handleMenu обрабатывает запрос меню
func (h *MessageHandler) handleMenu(bot *tgbotapi.BotAPI, chatID int64) error {
	text := "📋 Главное меню:\n\n" +
		"Доступные разделы:\n" +
		"• Профиль - информация о вас\n" +
		"• Настройки - настройки бота\n" +
		"• Меню - это сообщение\n" +
		"• Курсы - список доступных курсов"

	reply := tgbotapi.NewMessage(chatID, text)
	// Используем inline-клавиатуру вместо reply-клавиатуры
	kb := keyboard.NewMainMenuInlineKeyboard()
	kb = keyboard.AddBackButton(kb)
	reply.ReplyMarkup = &kb
	_, err := bot.Send(reply)
	return err
}

// handleHideKeyboard скрывает reply-клавиатуру
func (h *MessageHandler) handleHideKeyboard(bot *tgbotapi.BotAPI, chatID int64) error {
	text := "Привет! Я тестовый бот на Go.\n\n" +
		"Я могу помочь вам с различными задачами.\n\n" +
		"Доступные команды:\n" +
		"/start - начать работу\n" +
		"/help - помощь\n" +
		"/info - информация о вас"

	reply := tgbotapi.NewMessage(chatID, text)
	// Убираем клавиатуру
	hideKeyboard := tgbotapi.NewRemoveKeyboard(true)
	reply.ReplyMarkup = hideKeyboard

	_, err := bot.Send(reply)
	return err
}
