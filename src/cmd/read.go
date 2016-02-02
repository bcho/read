package main

import (
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/telegram-bot-api.v1"
)

type Record struct {
	Content string
	At      time.Time
}

func newRecord(content string) *Record {
	return &Record{content, time.Now()}
}

type RecordManager struct {
	Records []*Record

	contentChan chan string
	quitChan    chan struct{}
}

func newRecordManager() *RecordManager {
	return &RecordManager{
		contentChan: make(chan string),
		quitChan:    make(chan struct{}),
	}
}

func (r RecordManager) RecordContent(content string) {
	r.contentChan <- content
}

func (r *RecordManager) Start() {
	go r.start()
}

func (r *RecordManager) start() {
	for {
		select {
		case content := <-r.contentChan:
			log.Printf("Received new content: %s", content)
			r.Records = append(r.Records, newRecord(content))
		case <-r.quitChan:
			return
		}
	}
}

func (r RecordManager) Quit() {
	r.quitChan <- struct{}{}
}

func main() {
	manager := newRecordManager()
	manager.Start()

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

		manager.RecordContent(update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Roger that!")
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
