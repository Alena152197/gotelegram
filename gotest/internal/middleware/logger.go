package middleware

import (
    "log"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// LogCommand логирует команду перед обработкой
func LogCommand(msg *tgbotapi.Message) {
    user := msg.From
    command := msg.Command()
    
    log.Printf(
        "[%s] Команда /%s от пользователя %s (ID: %d) в чате %d",
        time.Now().Format("2006-01-02 15:04:05"),
        command,
        user.UserName,
        user.ID,
        msg.Chat.ID,
    )
}

// LogMessage логирует текстовое сообщение
func LogMessage(msg *tgbotapi.Message) {
    user := msg.From
    
    log.Printf(
        "[%s] Сообщение от пользователя %s (ID: %d): %s",
        time.Now().Format("2006-01-02 15:04:05"),
        user.UserName,
        user.ID,
        msg.Text,
    )
}