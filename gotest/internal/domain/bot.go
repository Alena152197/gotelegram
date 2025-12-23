package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Bot представляет экземпляр бота со всеми зависимостями
type Bot struct {
    API    *tgbotapi.BotAPI  // API для работы с Telegram
    Config *BotConfig         // Конфигурация бота
}

// BotConfig содержит настройки бота
type BotConfig struct {
    Token    string
    Debug    bool
    AdminIDs []int64
}