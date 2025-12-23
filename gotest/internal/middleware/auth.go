package middleware

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsAdmin проверяет, является ли пользователь администратором
func IsAdmin(userID int64, adminIDs []int64) bool {
    for _, adminID := range adminIDs {
        if userID == adminID {
            return true
        }
    }
    return false
}

// RequireAdmin проверяет права доступа и отправляет сообщение, если пользователь не админ
func RequireAdmin(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, adminIDs []int64) bool {
    userID := msg.From.ID
    
    if !IsAdmin(userID, adminIDs) {
        fmt.Println(userID, adminIDs)
        reply := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды.")
        bot.Send(reply)
        return false
    }
    
    return true
}