package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendMessage(uid int, token, text string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	bot.Debug = getEnvAsBool("DEBUG", false)
	msg := tgbotapi.NewMessage(int64(uid), text)
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}
