package main

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/config"
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

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Bot.Timeout
	updates := bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		handleUpdate(bot, update, cfg)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg *config.Config) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	if msg.IsCommand() {
		handleCommand(bot, msg, cfg)
		return
	}

	// Проверяем, является ли сообщение числом
	if num, err := strconv.ParseFloat(msg.Text, 64); err == nil {
		square := num * num
		response := fmt.Sprintf("Квадрат числа %g равен %g", num, square)
		sendMessage(bot, chatID, response)
	} else {
		sendMessage(bot, chatID, "Вы написали: "+msg.Text)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, cfg *config.Config) {
	command := msg.Command()

	switch command {
	case "start":
		handleStartCommand(bot, msg)
	case "help":
		handleHelpCommand(bot, msg)
	case "info":
		handleInfoCommand(bot, msg)
	case "settings":
		handleSettingsCommand(bot, msg, cfg)
	case "echo":
		handleEchoCommand(bot, msg)
	default:
		handleUnknownCommand(bot, msg)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "Привет! Я тестовый бот на Go.\n\n" +
		"Доступные команды:\n" +
		"/start - начать работу\n" +
		"/help - помощь\n" +
		"/info - информация о вас\n" +
		"/settings - настройки бота (только для администраторов)\n" +
		"/echo <текст> - повторить текст"

	sendMessage(bot, msg.Chat.ID, text)
}

func handleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "Список доступных команд:\n\n" +
		"/start - начать работу с ботом\n" +
		"/help - показать это сообщение\n" +
		"/info - получить информацию о себе\n" +
		"/settings - настройки бота (только для администраторов)\n" +
		"/echo <текст> - повторить указанный текст\n\n" +
		"Текстовые команды:\n" +
		"подписка - информация о подписке"

	sendMessage(bot, msg.Chat.ID, text)
}

func handleInfoCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	user := msg.From
	chatID := msg.Chat.ID

	info := "Информация о вас:\n\n"
	info += fmt.Sprintf("ID: %d\n", user.ID)
	info += fmt.Sprintf("Имя: %s\n", user.FirstName)

	if user.LastName != "" {
		info += fmt.Sprintf("Фамилия: %s\n", user.LastName)
	}

	if user.UserName != "" {
		info += fmt.Sprintf("Username: @%s\n", user.UserName)
	}

	info += fmt.Sprintf("Язык: %s\n", user.LanguageCode)
	info += fmt.Sprintf("Бот: %v\n", user.IsBot)

	sendMessage(bot, chatID, info)
}

func handleSettingsCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, cfg *config.Config) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	// Проверяем, является ли пользователь администратором
	isAdmin := false
	for _, adminID := range cfg.Bot.AdminIDs {
		if adminID == userID {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		// Формируем сообщение с настройками
		text := "Настройки бота:\n"
		text += fmt.Sprintf("Режим отладки: %v\n", cfg.Bot.Debug)
		text += fmt.Sprintf("Таймаут: %d секунд", cfg.Bot.Timeout)
		sendMessage(bot, chatID, text)
	} else {
		// Пользователь не администратор
		text := "Эта команда доступна только администраторам"
		sendMessage(bot, chatID, text)
	}
}

func handleEchoCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	// Получаем аргументы команды (все что после /echo)
	args := msg.CommandArguments()
	// Проверяем, есть ли аргументы
	if args == "" {
		// Аргументов нет - просим указать текст
		text := "Пожалуйста, укажите текст для повторения. Использование: /echo <текст>"
		sendMessage(bot, chatID, text)
	} else {
		// Аргументы есть - повторяем их
		sendMessage(bot, chatID, args)
	}
}

func handleUnknownCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "Неизвестная команда. Используйте /help для списка команд."
	sendMessage(bot, msg.Chat.ID, text)
}

// sendMessage безопасно отправляет сообщение
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
	}
}
