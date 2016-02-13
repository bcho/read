package robot

import (
	"fmt"
	"time"

	"brain"

	tgbot "gopkg.in/telegram-bot-api.v1"
)

type Robot interface {
	Start() error
	SetCommand(command, argument string)
	Response(message tgbot.Message) tgbot.MessageConfig
}

type robot struct {
	currCommand          string
	currCommanadArgument string

	articles  brain.Brain
	bookmarks brain.Brain
}

func NewRobot() *robot {
	return &robot{
		articles:  brain.NewBrain(),
		bookmarks: brain.NewBrain(),
	}
}

func (r *robot) Start() error {
	r.SetIdle()

	return nil
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

	// TODO extract & validate link format
	link := r.currCommanadArgument
	err := r.articles.Remember(msg.Time(), link, link)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Copy that! New link %s added.", link)
}

func (r *robot) responseStats(_ tgbot.Message) string {
	defer r.SetIdle()
	return "stats"
}

func (r *robot) responseBookmark(msg tgbot.Message) string {
	defer r.SetIdle()

	// TODO extract & validate link
	link := r.currCommanadArgument
	err := r.bookmarks.Remember(msg.Time(), link, link)
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
