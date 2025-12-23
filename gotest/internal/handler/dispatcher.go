package handler

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Dispatcher управляет обработчиками команд
type Dispatcher struct {
	handlers map[string]Handler // Карта: команда -> обработчик
}

// NewDispatcher создаёт новый диспетчер
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

// Register регистрирует обработчик
func (d *Dispatcher) Register(handler Handler) {
	command := handler.Command()
	d.handlers[command] = handler
	log.Printf("Зарегистрирован обработчик команды /%s", command)
}

// HandleCommand обрабатывает команду, направляя её к соответствующему обработчику
func (d *Dispatcher) HandleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	command := msg.Command()

	// Ищем обработчик для команды
	handler, exists := d.handlers[command]
	if !exists {
		// Обработчик не найден — отправляем сообщение о неизвестной команде
		return d.handleUnknownCommand(bot, msg)
	}

	// Вызываем обработчик
	err := handler.Handle(bot, msg)
	if err != nil {
		log.Printf("Ошибка обработки команды /%s: %v", command, err)
		return err
	}

	return nil
}

// handleUnknownCommand обрабатывает неизвестные команды
func (d *Dispatcher) handleUnknownCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	chatID := msg.Chat.ID
	text := "Неизвестная команда. Используйте /help для списка доступных команд."

	reply := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(reply)
	return err
}
