package telegram

import (
	"fmt"
	"log"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil{
		return nil, err
	}

	log.Printf(
		"Telegram bot authorized as @%s",
		bot.Self.UserName,
	)

	return &Bot{
		bot: bot,
	}, nil
}

func (b *Bot) SendMessage(
	chatId int64,
	text string,
) error {

	message:= tgbotapi.NewMessage(chatId, text)

	_, err := b.bot.Send(message)

	return err
}

func (b *Bot) String()string{
	return fmt.Sprintf("Telegram bot as @%s", b.bot.Self.UserName)
}


func (b *Bot) Start(
	ctx context.Context,
	handler func(context.Context, tgbotapi.Update),
) {

	config := tgbotapi.NewUpdate(0)
	config.Timeout = 30

	updates := b.bot.GetUpdatesChan(config)

	go func() {

		for {
			select {

			case update := <-updates:
				handler(ctx, update)

			case <-ctx.Done():
				b.bot.StopReceivingUpdates()
				return
			}
		}
	}()

	log.Println("Telegram bot started")
}