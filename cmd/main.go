package main

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"cibus-coupon-telegram-bot/internal/db"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	EnvVarTGBotToken     = "TG_BOT_TOKEN"
	EnvVarAllowedUserIDs = "ALLOWED_USER_IDS"
	EnvVarCouponsBucket  = "COUPONS_BUCKET"
)

func main() {
	// Load conf from env
	env := make(map[string]string)
	for _, name := range []string{
		EnvVarTGBotToken,
		EnvVarAllowedUserIDs,
		EnvVarCouponsBucket,
	} {
		val, found := os.LookupEnv(name)
		if !found {
			log.Panicf("required env var %q wasn't found", name)
		}
		env[name] = val
	}

	// Init DB client
	//dbClient, err := db.NewLocalDBClient() //- Uncomment for testing with local FS as db
	dbClient, err := db.NewS3Client(env[EnvVarCouponsBucket])
	if err != nil {
		log.Panic(err)
	}

	// Init telegram bot api
	bot, err := tgbotapi.NewBotAPI(env[EnvVarTGBotToken])
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on bot %s", bot.Self.UserName)

	// Config bot & updates
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Process arriving updates
	// Respond only to updates originated from authorized users (i.e., ALLOWED_USER_IDS)
	allowedUserIDs := strings.Split(env[EnvVarAllowedUserIDs], ",")
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
					break
				}
			}

			// Drop updates from unauthorized users
			// In debug mode also reply with a proper message
			if !authorized {
				log.Printf("Skipping update: user %q (ID: %s) is not authorized", sender, senderID)
				if bot.Debug {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot")
					msg.ReplyToMessageID = update.Message.MessageID
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

					// Notify the user if there are no available coupons
					if len(coupons) == 0 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, there are no available coupons")
						msg.ReplyToMessageID = update.Message.MessageID
						if _, err := bot.Send(msg); err != nil {
							log.Printf("Failed to reply to user %q on message id %q", sender, update.Message.MessageID)
						}
						continue
					}

					// Build a button keyboard where each coupon ID is a button
					var keyboardButtonRows [][]tgbotapi.KeyboardButton
					for _, coupon := range coupons {
						keyboardButtonRows = append(
							keyboardButtonRows,
							tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("/use %s", coupon.String()))),
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
					// Parse 'use' command and extract coupon ID
					var couponID, errMsgText string
					cmdArgs := strings.Split(update.Message.Text, " ")
					if len(cmdArgs) >= 2 {
						couponID = cmdArgs[1]
					} else {
						log.Printf("No coupon ID was given")
						errMsgText = "Please specify a coupon (/use <coupon-id>)"
					}

					// 'Use' given coupon
					if errMsgText == "" {
						if err := dbClient.Use(couponID); err != nil {
							switch err {
							case db.ErrCouponAlreadyUsed:
								errMsgText = "This coupon was already used"
							case db.ErrCouponNotExist:
								errMsgText = "There is no such coupon"
							default:
								log.Printf("Failed marking coupon %q as used", couponID)
								errMsgText = "Something went wrong. Please try again"
							}
						}
					}

					// Generate coupon's barcode
					// In case of an error, fallback to regular text message with the coupon ID
					var barcodePath, useCouponMsg string
					if errMsgText == "" {
						useCouponMsg = fmt.Sprintf("Using %s", coupon.ReadableCouponID(couponID))
						barcodePath, err = coupon.GenerateBarcodeFile(couponID)
						if err != nil {
							log.Printf("Failed creating barcode for coupon %q", couponID)
							errMsgText = useCouponMsg
						}
					}

					// Construct the message to reply with
					var photoMsg tgbotapi.PhotoConfig
					var textMsg tgbotapi.MessageConfig
					var msg tgbotapi.Chattable
					if errMsgText == "" {
						photoMsg = tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath(barcodePath))
						photoMsg.Caption = useCouponMsg
						photoMsg.ReplyToMessageID = update.Message.MessageID
						photoMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true) // Close keyboard opened in /list
						msg = photoMsg
					} else {
						textMsg = tgbotapi.NewMessage(update.Message.Chat.ID, errMsgText)
						textMsg.ReplyToMessageID = update.Message.MessageID
						textMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true) // Close keyboard opened in /list
						msg = textMsg
					}

					// Reply
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to reply to user %q on message id %q", sender, update.Message.MessageID)
					}
				}
			}
		}
	}
}
