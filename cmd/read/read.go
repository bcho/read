package main

import (
	"log"
	"os"

	"github.com/bcho/read/robot"
	_ "github.com/joho/godotenv/autoload"
	tgbotapi "gopkg.in/telegram-bot-api.v1"
)

func main() {
	robot := robot.NewRobot()
	if err := robot.Start(); err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		bot.Send(robot.Response(update.Message))
	}
}
