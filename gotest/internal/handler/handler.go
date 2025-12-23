package handler

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Handler — интерфейс для обработчиков команд
type Handler interface {
    Handle(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error
    Command() string  // Возвращает команду, которую обрабатывает этот обработчик
}