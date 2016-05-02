package robot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bcho/read/brain"
	"github.com/bcho/read/publisher"
	"github.com/bcho/timespan"
	"github.com/mvdan/xurls"
	tgbot "gopkg.in/telegram-bot-api.v1"
)

var (
	dumpInterval = time.Duration(30) * time.Second
)

type Robot interface {
	Start() error
	Stop()
	SetCommand(command, argument string)
	Response(message tgbot.Message) tgbot.MessageConfig
}

type robot struct {
	currCommand          string
	currCommanadArgument string

	articles  brain.Brain
	bookmarks brain.Brain
	publisher publisher.Publisher

	stop chan struct{}
}

func NewRobot() *robot {
	return &robot{
		articles:  brain.NewBrain(),
		bookmarks: brain.NewBrain(),

		publisher: publisher.NewMediumPublisher(os.Getenv("MEDIUM_TOKEN")),

		stop: make(chan struct{}),
	}
}

func (r *robot) Start() error {
	defer r.SetIdle()

	r.restoreBrain()

	go func() {
		dumpTimeout := time.Tick(dumpInterval)
		for {
			select {
			case <-r.stop:
				r.stop <- struct{}{}
				return
			case <-dumpTimeout:
				r.dumpBrain()
			}
		}
	}()

	return nil
}

func (r robot) Stop() {
	r.stop <- struct{}{}
	<-r.stop
}

func (r *robot) SetCommand(command, argument string) {
	r.currCommand = command
	r.currCommanadArgument = argument
}

func (r *robot) SetIdle() {
	r.SetCommand(Idle, "")
}

func (r *robot) Response(message tgbot.Message) tgbot.MessageConfig {
	if message.IsCommand() {
		r.SetCommand(message.Command(), message.CommandArguments())
	}

	var response string
	switch r.currCommand {
	case Idle:
		response = r.responseIdle(message)
	case Read:
		response = r.responseRead(message)
	case Stats:
		response = r.responseStats(message)
	case Bookmark:
		response = r.responseBookmark(message)
	case Random:
		response = r.responseRandom(message)
	case Publish:
		response = r.responsePublish(message)
	}

	reply := tgbot.NewMessage(message.Chat.ID, response)
	reply.ReplyToMessageID = message.MessageID

	return reply
}

func (r *robot) responseIdle(_ tgbot.Message) string {
	return "What can I do for you?"
}

func (r *robot) responseRead(msg tgbot.Message) string {
	defer r.SetIdle()

	link := extractFirstLink(r.currCommanadArgument)
	if link == "" {
		return "Oops, can't find any links."
	}

	err := r.articles.Remember(msg.Time(), link, r.currCommanadArgument)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Copy that! New link %s added.", link)
}

func (r *robot) responseStats(_ tgbot.Message) string {
	defer r.SetIdle()

	daysBefore := 7
	fmt.Sscanf(r.currCommanadArgument, "%d", &daysBefore)
	statsDuration := time.Duration(-daysBefore*24) * time.Hour
	statsSpan := timespan.New(time.Now(), statsDuration)

	things := r.articles.GetInPeriod(statsSpan)
	return fmt.Sprintf(
		"You read %d article(s) during %s ~ %s:\n\n%s",
		len(things),
		statsSpan.Start().Format("2006-01-02"),
		statsSpan.End().Format("2006-01-02"),
		strings.Join(things, "\n-------------------------------\n\n"),
	)
}

func (r *robot) responseBookmark(msg tgbot.Message) string {
	defer r.SetIdle()

	link := extractFirstLink(r.currCommanadArgument)
	if link == "" {
		return "Oops, can't find any links."
	}

	err := r.bookmarks.Remember(msg.Time(), link, r.currCommanadArgument)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Roger that! New link %s added.", link)
}

func (r *robot) responseRandom(_ tgbot.Message) string {
	defer r.SetIdle()

	randomLink := ""
	randomKey := ""

	err := r.bookmarks.Each(func(_ time.Time, key, link string) error {
		randomLink = link
		randomKey = key

		return brain.EachBreak
	})
	if err != nil && err != brain.EachBreak {
		return err.Error()
	}

	if randomLink == "" {
		return "No more bookmarks, nice!"
	}

	if err := r.bookmarks.Forget(randomKey); err != nil {
		return err.Error()
	}

	return randomLink
}

func (r *robot) responsePublish(_ tgbot.Message) string {
	defer r.SetIdle()

	daysBefore := 7
	fmt.Sscanf(r.currCommanadArgument, "%d", &daysBefore)
	statsDuration := time.Duration(-daysBefore*24) * time.Hour
	statsSpan := timespan.New(time.Now(), statsDuration)
	things := r.articles.GetInPeriod(statsSpan)

	postUrl, err := r.publisher.Publish(statsSpan, things)
	if err != nil {
		return err.Error()
	}
	return postUrl
}

func (r robot) dumpBrain() {
	var (
		err       error
		articles  []map[string]string
		bookmarks []map[string]string
	)

	err = r.articles.Each(func(at time.Time, key, thing string) error {
		articles = append(articles, map[string]string{
			"at":    at.Format(time.RFC3339),
			"key":   key,
			"thing": thing,
		})

		return nil
	})
	if err != nil {
		log.Printf("dump brain failed: %v", err)
		return
	}

	err = r.bookmarks.Each(func(at time.Time, key, thing string) error {
		bookmarks = append(bookmarks, map[string]string{
			"at":    at.Format(time.RFC3339),
			"key":   key,
			"thing": thing,
		})

		return nil
	})
	if err != nil {
		log.Printf("dump brain failed: %v", err)
		return
	}

	data := map[string][]map[string]string{
		"articles":  articles,
		"bookmarks": bookmarks,
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("dump brain failed: %v", err)
		return
	}

	if err := ioutil.WriteFile(dumpFile(), b, 0644); err != nil {
		log.Printf("dump brain failed: %v", err)
		return
	}

	log.Printf("dump brain to %s finished", dumpFile())
}

func (r *robot) restoreBrain() {
	file, err := os.Open(dumpFile())
	if err != nil {
		log.Printf("restore brain failed: %v", err)
		return
	}

	data := make(map[string][]map[string]string)
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Printf("restore brain failed: %v", err)
		return
	}

	if _, present := data["articles"]; !present {
		log.Printf("courrpted dump file")
		return
	}
	if _, present := data["bookmarks"]; !present {
		log.Printf("courrpted dump file")
		return
	}

	restoredArticles := 0
	restoredBookmarks := 0
	for _, thing := range data["articles"] {
		at, err := time.Parse(time.RFC3339, thing["at"])
		if err != nil {
			log.Printf("courrpted dump file: %s", thing["at"])
			continue
		}
		if err := r.articles.Remember(at, thing["key"], thing["thing"]); err != nil {
			log.Printf("courrpted dump file")
			continue
		}
		restoredArticles += 1
	}
	for _, thing := range data["bookmarks"] {
		at, err := time.Parse(time.RFC3339, thing["at"])
		if err != nil {
			log.Printf("courrpted dump file: %s", thing["at"])
			continue
		}
		if err := r.bookmarks.Remember(at, thing["key"], thing["thing"]); err != nil {
			log.Printf("courrpted dump file")
			continue
		}
		restoredBookmarks += 1
	}

	log.Printf(
		"restored %d articles, %d bookmarks from %s",
		restoredArticles,
		restoredBookmarks,
		dumpFile(),
	)
}

func extractFirstLink(content string) string {
	return xurls.Strict.FindString(content)
}

func dumpFile() string {
	f := os.Getenv("DUMP")
	if f == "" {
		f = "./.dump"
	}
	return f
}
