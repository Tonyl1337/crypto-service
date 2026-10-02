package telegram

import (
	"github.com/Tonyl1337/crypto-service/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💰 Курсы"),
			tgbotapi.NewKeyboardButton("🔔 Подписаться"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Подписки"),
			tgbotapi.NewKeyboardButton("ℹ️ Помощь"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func ratesKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("BTC"),
			tgbotapi.NewKeyboardButton("ETH"),
			tgbotapi.NewKeyboardButton("SOL"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("TON"),
			tgbotapi.NewKeyboardButton("DOGE"),
			tgbotapi.NewKeyboardButton("ADA"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔎 Другая монета"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⬅️ Назад"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func subscriptionCoinKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("BTC"),
			tgbotapi.NewKeyboardButton("ETH"),
			tgbotapi.NewKeyboardButton("SOL"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("TON"),
			tgbotapi.NewKeyboardButton("DOGE"),
			tgbotapi.NewKeyboardButton("ADA"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔎 Другая монета"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func intervalKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1 мин"),
			tgbotapi.NewKeyboardButton("5 мин"),
			tgbotapi.NewKeyboardButton("10 мин"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("30 мин"),
			tgbotapi.NewKeyboardButton("60 мин"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func subscriptionsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑 Удалить подписку"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑 Удалить все"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⬅️ Назад"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func subscriptionDeleteKeyboard(
	subscriptions []domain.Subscription,
) tgbotapi.ReplyKeyboardMarkup {
	rows := make([][]tgbotapi.KeyboardButton, 0, len(subscriptions)+1)

	for _, subscription := range subscriptions {
		if !subscription.Enabled {
			continue
		}

		rows = append(
			rows,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(subscription.Symbol),
			),
		)
	}

	rows = append(
		rows,
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.ResizeKeyboard = true

	return keyboard
}
