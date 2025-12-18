package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/internal/config"
)

// handleUpdate обрабатывает одно обновление от Telegram
func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Проверяем, есть ли сообщение
	if update.Message == nil {
		return // Если сообщения нет — игнорируем обновление
	}

	// Получаем сообщение и чат
	msg := update.Message
	chatID := msg.Chat.ID

	// Проверяем, является ли сообщение командой
	if msg.IsCommand() {
		handleCommand(bot, msg)
		return
	}

	if msg.Text == "подписка" {
		reply := tgbotapi.NewMessage(chatID, "Если вы хотите купить подписку неапишите администратору @MyAppBook")
		bot.Send(reply)
	} else {
		reply := tgbotapi.NewMessage(chatID, "Вы написали: "+msg.Text)
		bot.Send(reply)
	}

	// Если это не команда — отвечаем эхом
	reply := tgbotapi.NewMessage(chatID, "Вы написали: "+msg.Text)
	bot.Send(reply)
}

// handleCommand обрабатывает команды бота
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	command := msg.Command()

	// Обрабатываем разные команды
	switch command {
	case "start":
		handleStartCommand(bot, chatID)
	case "help":
		handleHelpCommand(bot, chatID)
	default:
		handleUnknownCommand(bot, chatID)
	}
}

// handleStartCommand обрабатывает команду /start
func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Привет! Я тестовый бот на Go.\n\n" +
		"Доступные команды:\n" +
		"/start - начать работу\n" +
		"/help - помощь"

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// handleHelpCommand обрабатывает команду /help
func handleHelpCommand(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Это справочная информация.\n\n" +
		"Бот создан с помощью библиотеки go-telegram-bot-api.\n" +
		"Исходный код: https://github.com/go-telegram-bot-api/telegram-bot-api"

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// handleUnknownCommand обрабатывает неизвестные команды
func handleUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Неизвестная команда. Используйте /help для списка команд."

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Создаём экземпляр бота с токеном
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	// Включаем режим отладки (если нужно)
	bot.Debug = cfg.Bot.Debug

	// Выводим информацию о боте
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Создаём канал для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Bot.Timeout

	// Получаем канал обновлений
	updates := bot.GetUpdatesChan(u)

	// Обрабатываем каждое обновление
	for update := range updates {
		// Обрабатываем обновление
		handleUpdate(bot, update)
	}
}
