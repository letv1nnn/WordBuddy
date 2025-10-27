package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/letv1nnn/WordBoddy/pkg/db"
)

type Flag int

const (
	NoneFlag Flag = iota
	StartFlag
	TranslateFlag
)

type Bot struct {
	bot   *tgbotapi.BotAPI
	token string
	db    *db.Storage
	flag  Flag
}

func NewBot(bot *tgbotapi.BotAPI, token string, db *db.Storage) *Bot {
	return &Bot{bot, token, db, NoneFlag}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	updates, err := b.initUpdatesChannel()
	if err != nil {
		return err
	}

	if err := b.registerCommands(); err != nil {
		return err
	}

	b.handleUpdates(updates)

	return nil
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if b.flag != 0 {
			b.handleFlag(*update.Message)
			continue
		}

		if update.Message.IsCommand() {
			if err := b.handleCommand(*update.Message); err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				b.bot.Send(msg)
			}
			continue
		}

		b.handleMessage(*update.Message)
	}
}

func (b *Bot) initUpdatesChannel() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.bot.GetUpdatesChan(u), nil
}

func (b *Bot) registerCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start bot"},
		{Command: "help", Description: "Show help message"},
		{Command: "add", Description: "Add word"},
		{Command: "me", Description: "Get yout data"},
	}

	_, err := b.bot.Request(tgbotapi.NewSetMyCommands(commands...))
	return err
}
