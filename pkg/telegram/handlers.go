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
		if b.checkIfRegistered(command) {
			msg.Text = "✅ Successfully signed in"
		} else {
			msg.Text = "Hi! 🌍\nPlease choose the language you know and the language you want to learn, e.g., \"english russian\":"
			b.flag = StartFlag
		}
	case "help":
		// file :=
		// tgbotapi.NewPhoto(command.Chat.ID, file)
		msg.Text = "WordBuddy is a Telegram bot that helps people learn foreign language vocabulary. It provides a word storage and organizes quizes. You can review the availables commands by pressing the blue menu button at the bottom left corner."
	case "me":
		userInfo, err := b.getUserInfo(int(command.From.ID))
		if err != nil {
			return err
		}
		msg.Text = *userInfo
	case "add":
		msg.Text = "Enter one or multiple words separated by comma that you want to translate and save."
		b.flag = TranslateFlag
	case "list":
		if err := b.handleList(command); err != nil {
			return err
		}
	case "quiz":
		if err := b.handleQuiz(command); err != nil {
			return err
		}
	default:
		msg.Text = "I don't know this command"
	}
	b.bot.Send(msg)
	return nil
}

func (b *Bot) handleFlag(message tgbotapi.Message) {
	switch b.flag {
	case StartFlag:
		user := b.handleStart(message)
		if user == nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to Sign Up")
			b.bot.Send(msg)
			return
		}
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
	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Successfully signed up"))
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

func (b *Bot) handleList(message tgbotapi.Message) error {
	user, err := b.db.Get(int(message.From.ID))
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to get the data from the database")
		b.bot.Send(msg)
		return err
	}

	text := "You are storing " + fmt.Sprint(len(user.Words)) + " words."
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.bot.Send(msg)

	for _, word := range user.Words {
		text := string(word.Original) + " - " + string(word.Translated[0])
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.bot.Send(msg)
	}

	return nil
}

func (b *Bot) handleQuiz(message tgbotapi.Message) error {
	user, err := b.db.Get(int(message.From.ID))
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to get the data from the database")
		b.bot.Send(msg)
		return err
	}
	langFrom, langTo := user.LanguageFrom, user.LanguageTo
	maxWords := 0

	for _, word := range user.Words {
		prompt := "Words" + string(word.Original) + " - " + string(word.Translated[0]) + "\n" +
			"Translation from" + langFrom + ", Translation to" + langTo

		response, err := apirequests.MakeOllamaRequest(prompt)
		if err != nil {
			return err
		}

		lines := strings.Split(*response, "\n")
		var questionLines []string
		var correctAnswer string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Answer:") {
				correctAnswer = strings.TrimSpace(strings.TrimPrefix(line, "Answer:"))
			} else {
				questionLines = append(questionLines, line)
			}
		}

		questionText := strings.Join(questionLines, "\n")

		options := []string{"A", "B", "C", "D"}
		var buttons [][]tgbotapi.InlineKeyboardButton

		for i := 0; i < len(options); i += 2 {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(options[i], options[i]+"|"+correctAnswer),
			}
			if i+1 < len(options) {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(options[i+1], options[i+1]+"|"+correctAnswer))
			}
			buttons = append(buttons, row)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		msg := tgbotapi.NewMessage(message.Chat.ID, questionText)
		msg.ReplyMarkup = keyboard
		b.bot.Send(msg)

		maxWords++
		if maxWords > 20 {
			break
		}
	}

	return nil
}

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	parts := strings.Split(query.Data, "|")
	selected := parts[0]
	correct := parts[1]

	var text string
	if selected == correct {
		text = "✅ Correct!"
	} else {
		text = "❌ Wrong! Correct answer: " + correct
	}

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	b.bot.Send(msg)

	callback := tgbotapi.NewCallback(query.ID, "")
	b.bot.Request(callback)
}
