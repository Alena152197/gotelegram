package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/middleware"
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

	// Создаём диспетчер обработчиков
	dispatcher := handler.NewDispatcher()

	// Регистрируем обработчики команд
	dispatcher.Register(handler.NewStartHandler())
	dispatcher.Register(handler.NewHelpHandler())
	dispatcher.Register(handler.NewInfoHandler())
	dispatcher.Register(handler.NewAdminHandler(cfg.Bot.AdminIDs))

	// Создаём обработчик обычных сообщений
	messageHandler := handler.NewMessageHandler()

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Bot.Timeout
	updates := bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		handleUpdate(bot, dispatcher, messageHandler, update)
	}
}

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
