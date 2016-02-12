package robot

import (
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
}

func NewRobot() *robot {
	return &robot{}
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
	case Random:
		response = r.responseStats(message)
	}

	reply := tgbot.NewMessage(message.Chat.ID, response)
	reply.ReplyToMessageID = message.MessageID

	return reply
}

func (r *robot) responseIdle(_ tgbot.Message) string {
	return "idle"
}

func (r *robot) responseRead(_ tgbot.Message) string {
	defer r.SetIdle()
	return "read"
}

func (r *robot) responseStats(_ tgbot.Message) string {
	defer r.SetIdle()
	return "stats"
}

func (r *robot) responseRandom(_ tgbot.Message) string {
	defer r.SetIdle()
	return "random"
}
