package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/grovesbs/readingls/orm"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	conn := orm.New(db)

	bot, err := telegram.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := telegram.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			msg := telegram.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "Hey! Welcome to Reading List. I store any links you send me and can relay them back to you /mylist.\nMore features coming soon!"
			case "mylist":
				links, err := conn.GetLinks(update.Message.From.UserName)
				if err != nil {
					log.Panic(err)
				}
				msg.Text = ""
				for i, e := range links {
					msg.Text += fmt.Sprintf("%d. %s\n", i+1, e.Url)
				}
			}
			if _, err := bot.Send(msg); err != nil {
				log.Panicln(err)
			}
			continue
		}

		// If we got a message
		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		userURL, err := url.ParseRequestURI(update.Message.Text)
		if err != nil {
			msg := telegram.NewMessage(update.Message.Chat.ID, "Sorry, I can only save links...")
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
			continue
		}
		if err := conn.InsertURL(userURL, update.Message.From.UserName); err != nil {
			msg := telegram.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error saving %s. Please try again", userURL))
			bot.Send(msg)
			continue
		}

		msg := telegram.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Saved %s", userURL))
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := bot.Send(msg); err != nil {
			log.Panicln(err)
		}

	}
}
