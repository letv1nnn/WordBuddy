package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	apirequests "github.com/letv1nnn/WordBoddy/pkg/api-requests"
	"github.com/letv1nnn/WordBoddy/pkg/db"
)

func (b *Bot) handleMessage(message tgbotapi.Message) {
	log.Printf("[%d] %s", message.Chat.ID, message.Text)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Do not make this chat dirty pls")
	b.bot.Send(msg)
}

func (b *Bot) handleCommand(command tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(command.Chat.ID, "")
	switch command.Command() {
	case "start":
		msg.Text = "Hi! 🌍\nPlease choose the language you know and the language you want to learn, e.g., \"english russian\":"
		b.flag = StartFlag
	case "help":
		msg.Text = "you've entered help command"
	case "me":
		userInfo, err := b.getUserInfo(int(command.From.ID))
		if err != nil {
			return err
		}
		msg.Text = *userInfo
	case "add":
		msg.Text = "Enter one or multiple words separated by comma that you want to translate and save."
		b.flag = TranslateFlag
	default:
		msg.Text = "I don't know this command"
	}
	_, err := b.bot.Send(msg)
	return err
}

func (b *Bot) handleFlag(message tgbotapi.Message) {
	switch b.flag {
	case StartFlag:
		user := b.handleStart(message)
		if err := b.db.Save(*user); err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to save the data to the database")
			b.bot.Send(msg)
		}
	case TranslateFlag:
		b.handleAdd(message)
	}
}

func (b *Bot) handleStart(message tgbotapi.Message) *db.User {
	langs := strings.Split(message.Text, " ")
	if len(langs) != 2 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Only two words are allowed, your original language and the language you want to lean e.g., \"english russian\". Try again.")
		msg.ReplyToMessageID = message.MessageID
		b.bot.Send(msg)
		return nil
	}
	words := make([]db.Word, 0)
	user := db.NewUser(int(message.From.ID), message.Chat.UserName, words, strings.ToLower(langs[1]), strings.ToLower(langs[0]))
	b.flag = NoneFlag
	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Successfully signed up"))
	return &user
}

func (b *Bot) handleAdd(message tgbotapi.Message) {
	b.flag = NoneFlag
	user, err := b.db.Get(int(message.From.ID))
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to get the data from the database")
		b.bot.Send(msg)
		b.flag = NoneFlag
		return
	}

	if user == nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ You are not registered yet. Use /start first."))
		b.flag = NoneFlag
		return
	}

	wordList := strings.Fields(message.Text)
	if len(wordList) == 0 {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Please provide some text to translate."))
		return
	}

	translated, err := apirequests.TranslateText(message.Text, user.LanguageFrom, user.LanguageTo)
	if err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Word translation failed"))
		return
	}

	for i, word := range wordList {
		msg := fmt.Sprintf("%s -> %s", word, translated[i])
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg))

		w := db.Word{
			Original:   word,
			Translated: []string{translated[i]},
		}
		user.Words = append(user.Words, w)
	}

	if err := b.db.Save(*user); err != nil {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to save words to database"))
		return
	}

	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ All words processed!"))
}
