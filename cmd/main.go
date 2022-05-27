package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/idanatias/cibus-coupons-telegram-bot/pkg/coupon"
	"github.com/idanatias/cibus-coupons-telegram-bot/pkg/db"
	"github.com/idanatias/cibus-coupons-telegram-bot/pkg/interfaces"

	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	EnvVarTGBotToken     = "TG_BOT_TOKEN"
	EnvVarAllowedUserIDs = "ALLOWED_USER_IDS"
	EnvVarCouponsBucket  = "COUPONS_BUCKET"
)

var (
	env      map[string]string
	dbClient interfaces.DB
	bot      *tgbotapi.BotAPI
)

// sendBot is a wrapper for sending messages to the bot
func sendBot(msg tgbotapi.Chattable) error {
	log.Printf("msg to send: %+v", msg)
	_, err := bot.Send(msg)
	return err
}

// handle requests arriving at the Lambda function, assuming these are telegram bot updates
func handle(ctx context.Context, rawUpdate map[string]interface{}) error {
	// Convert the json payload to an 'Update' object
	var update tgbotapi.Update
	updateBody := reflect.ValueOf(rawUpdate["body"]).String()
	if err := json.Unmarshal([]byte(updateBody), &update); err != nil {
		log.Printf("failed decoding json: %v", err)
		return err
	}

	// Only handle 'message' updates
	if update.Message == nil {
		log.Printf("Only handling message updates")
		return nil
	}

	// Check if user is authorized to use this bot
	var authorized bool
	allowedUserIDs := strings.Split(env[EnvVarAllowedUserIDs], ",")
	sender := update.Message.From.UserName
	senderID := strconv.FormatInt(update.Message.From.ID, 10)
	for _, allowedUserID := range allowedUserIDs {
		if allowedUserID == senderID {
			authorized = true
			break
		}
	}

	// Drop updates from unauthorized users
	if !authorized {
		log.Printf("Skipping update: user %q (ID: %s) is not authorized", sender, senderID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot")
		msg.ReplyToMessageID = update.Message.MessageID
		return sendBot(msg)
	}

	// Drop non-commands messages
	if !update.Message.IsCommand() {
		log.Printf("Skipping update: message %q is not a command", update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I only support commands")
		msg.ReplyToMessageID = update.Message.MessageID
		return sendBot(msg)
	}

	// Detect & handle command
	switch update.Message.Command() {
	case "list":
		{
			// Load coupons from DB
			coupons, err := dbClient.List()
			if err != nil {
				return fmt.Errorf("failed listing coupons: %v", err)
			}

			// Notify the user if there are no available coupons
			if len(coupons) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, there are no available coupons")
				msg.ReplyToMessageID = update.Message.MessageID
				return sendBot(msg)
			}

			// Build a button keyboard where each coupon is a button
			var keyboardButtonRows [][]tgbotapi.KeyboardButton
			for _, c := range coupons {
				keyboardButtonRows = append(
					keyboardButtonRows,
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("/use %s", c.String()))),
				)
			}
			couponsKeyboard := tgbotapi.NewReplyKeyboard(keyboardButtonRows...)

			// Reply with coupons keyboard
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select the one you want to use:")
			msg.ReplyMarkup = couponsKeyboard
			return sendBot(msg)
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
			var err error
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
			return sendBot(msg)
		}
	default:
		log.Printf("Command is not supported: %q", update.Message.Command())
	}
	return nil
}

func main() {
	var err error

	// Load conf from env
	env = make(map[string]string)
	for _, name := range []string{
		EnvVarTGBotToken,
		EnvVarAllowedUserIDs,
		EnvVarCouponsBucket,
	} {
		val, found := os.LookupEnv(name)
		if !found {
			log.Fatalf("required env var %q wasn't found", name)
		}
		env[name] = val
	}

	// Init DB client
	//dbClient, err := db.NewLocalDBClient() //- Uncomment for testing with local FS as db
	dbClient, err = db.NewS3Client(env[EnvVarCouponsBucket])
	if err != nil {
		log.Fatalf("failed to init db client: %v", err)
	}

	// Init telegram bot api
	bot, err = tgbotapi.NewBotAPI(env[EnvVarTGBotToken])
	if err != nil {
		log.Fatalf("failed to init telegram bot api: %v", err)
	}
	bot.Debug = true
	log.Printf("Authorized on bot %s", bot.Self.UserName)

	// Wait for and process requests arriving to the Lambda
	lambda.Start(handle)
}
