package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	channelId int64
	bot       *tgbotapi.BotAPI
}

func NewBot(token string, channelId int64) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:       bot,
		channelId: channelId,
	}, nil
}

func (bot *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				switch update.Message.Command() {
				case "help":
					msg.Text = bot.printHelp()
				case "channel":
					msg.Text = fmt.Sprintf("Channel ID: %d", update.Message.Chat.ID)
				case "check":
					msg = tgbotapi.NewMessage(bot.channelId, "If you received this message, everything is configured properly! :)")
				default:
					msg.Text = bot.printHelp()
				}

				if err := bot.RawSend(msg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}
		}
	}()
}

func (bot *Bot) printHelp() string {
	var b strings.Builder

	b.WriteString("Available commands:\n\n")
	b.WriteString("/help - Get this help\n")
	b.WriteString("/channel - Return the channel id\n")
	b.WriteString("/check - Send a message using the channel id from the configuration file (to verify it's working)\n")

	return b.String()
}

func (bot *Bot) RawSend(msg tgbotapi.MessageConfig) error {
	_, err := bot.bot.Send(msg)
	return err
}

func (bot *Bot) Send(text string) error {
	if bot.channelId == 0 {
		return fmt.Errorf("empty channel id, please use the /c command to obtain the value from the bot")
	}
	msg := tgbotapi.NewMessage(bot.channelId, text)
	_, err := bot.bot.Send(msg)
	return err
}
