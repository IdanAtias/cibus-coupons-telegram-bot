package main

import (
	"cibus-coupon-telegram-bot/internal/db"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Init telegram bot api using the bot token
	token, found := os.LookupEnv("TG_BOT_TOKEN")
	if !found {
		log.Panic("bot token wasn't found")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on bot %s", bot.Self.UserName)

	// Config bot & updates
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	allowedUserIDsStr, found := os.LookupEnv("ALLOWED_USER_IDS") // Bot will only respond to messages originated from these user IDs
	if !found {
		log.Panic("no allowed user IDs found")
	}
	allowedUserIDs := strings.Split(allowedUserIDsStr, ",")

	// Init DB client
	dbClient, err := db.NewLocalDBClient()
	if err != nil {
		log.Panic(err)
	}

	// Process arriving updates
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			// Check if user is authorized to use this bot
			var authorized bool
			sender := update.Message.From.UserName
			senderID := strconv.FormatInt(update.Message.From.ID, 10)
			for _, allowedUserID := range allowedUserIDs {
				if allowedUserID == senderID {
					authorized = true
				}
			}

			// Drop updates from unauthorized users
			// In debug mode also reply with a proper message
			if !authorized {
				log.Printf("Skipping update: user %q (ID: %s) is not authorized", sender, senderID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot")
				msg.ReplyToMessageID = update.Message.MessageID
				if bot.Debug {
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to reply to user %q on message id %q", sender, update.Message.MessageID)
					}
				}
				continue
			}

			// Drop non-commands messages and notify user
			if !update.Message.IsCommand() {
				log.Printf("Skipping update: message %q is not a command", update.Message.Text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I only support commands")
				msg.ReplyToMessageID = update.Message.MessageID
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Failed to reply to user %q on message id %q", sender, update.Message.MessageID)
				}
				continue
			}

			// Detect & handle commands
			switch update.Message.Command() {
			case "list":
				{
					// Load coupons from DB
					coupons, err := dbClient.List()
					if err != nil {
						log.Printf("Failed listing coupons, dropping")
						continue
					}

					// Build a button keyboard where each coupon ID is a button
					var keyboardButtonRows [][]tgbotapi.KeyboardButton
					for _, coupon := range coupons {
						couponStr := fmt.Sprintf("%s - %vILS - %s - %s", coupon.ID, coupon.Value, coupon.Vendor, time.Unix(1652904232, 0))
						keyboardButtonRows = append(
							keyboardButtonRows,
							tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("/use %s", couponStr))),
						)
					}
					couponsKeyboard := tgbotapi.NewReplyKeyboard(keyboardButtonRows...)

					// Reply with coupons keyboard
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = couponsKeyboard
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to reply to user %q on message id %q", sender, update.Message.MessageID)
					}
				}
			case "use":
				{
					log.Printf("DETECTED /USE")
				}
			}
		}
	}
}
