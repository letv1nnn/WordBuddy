package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/letv1nnn/WordBoddy/pkg/db"
	"github.com/letv1nnn/WordBoddy/pkg/telegram"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("NOTE: .env file not found, relying on environment variables")
	}
	token := os.Getenv("TELEGRAM_TOKEN")
	dbPath := os.Getenv("DB_PATH")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// initializing db
	storage, err := db.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := storage.InitTable(telegram.UsersTable + "\n" + telegram.WordsTable); err != nil {
		log.Fatal(err)
	}

	// wrapping my bot to my struct
	tgbot := telegram.NewBot(bot, token, storage)
	if err := tgbot.Start(); err != nil {
		log.Fatal(err)
	}
}
