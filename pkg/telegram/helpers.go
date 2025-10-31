package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) checkIfRegistered(message tgbotapi.Message) bool {
	user, err := b.db.Get(int(message.From.ID))
	return err == nil && user != nil
}
