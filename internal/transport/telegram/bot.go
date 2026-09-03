package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
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

	message := tgbotapi.NewMessage(chatId, text)

	_, err := b.bot.Send(message)

	return err
}

func (b *Bot) String() string {
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
			case update, ok := <-updates:
				if !ok {
					log.Println("Telegram updates channel closed")
					return
				}

				handler(ctx, update)

			case <-ctx.Done():
				b.bot.StopReceivingUpdates()
				log.Println("Telegram bot stopped")
				return
			}
		}
	}()

	log.Println("Telegram bot started")
}
