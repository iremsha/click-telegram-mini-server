package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mymmrac/telego"
)

func startBot(telegtamToken string) {
	log.Println("Starting bot...")
	bot, err := telego.NewBot(telegtamToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Println("Bot created successfully.")

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatalf("Failed to start receiving updates: %v", err)
	}
	log.Println("Bot is now receiving updates.")

	for update := range updates {
		if update.Message != nil {
			log.Printf("Received message from chat ID %d: %s", update.Message.Chat.ID, update.Message.Text)
			handleTelegramMessage(bot, update.Message)
		}
	}
}

func handleTelegramMessage(bot *telego.Bot, message *telego.Message) {
	chatID := telego.ChatID{
		ID: message.Chat.ID,
	}
	text := message.Text

	log.Printf("Handling message for chat ID %d: %s", chatID.ID, text)

	switch text {
	case "/start":
		response := "Welcome to Telegram Game Bot! Use /profile to see your stats."
		_, _ = bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
		log.Printf("Sent start message to chat ID %d", chatID.ID)

	case "/profile":
		telegramID := strconv.FormatInt(message.Chat.ID, 10)
		user, err := getUserByTelegramID(telegramID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				response := "Profile not found. Use /register to create a profile."
				_, _ = bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
				log.Printf("Profile not found for telegram ID %s, sent registration prompt.", telegramID)
			} else {
				log.Printf("Error retrieving user for telegram ID %s: %v", telegramID, err)
			}
			return
		}
		response := fmt.Sprintf("Name: %s\nPoints: %d", user.Name, user.Points)
		_, _ = bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
		log.Printf("Sent profile information to chat ID %d", chatID.ID)

	case "/register":
		telegramID := strconv.FormatInt(message.Chat.ID, 10)
		user := Player{
			TelegramID:  telegramID,
			Name:        message.From.FirstName,
			Points:      0,
			Energy:      10,
			LevelEnergy: 1,
			LevelX10:    1,
			LevelX100:   1,
			LevelX1000:  1,
		}
		err := insertOrUpdateUser(user)
		if err != nil {
			log.Printf("Error creating user for telegram ID %s: %v", telegramID, err)
			return
		}
		response := "Registration successful! Use /profile to see your stats."
		_, _ = bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
		log.Printf("Registered new user with telegram ID %s", telegramID)

	default:
		response := "Unknown command. Please use /profile or /register."
		_, _ = bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
		log.Printf("Sent unknown command message to chat ID %d", chatID.ID)
	}
}

func getUserByTelegramID(telegramID string) (Player, error) {
	log.Printf("Retrieving user for telegram ID: %s", telegramID)
	var user Player
	collection := db.Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"telegramId": telegramID}).Decode(&user)
	if err != nil {
		log.Printf("Error retrieving user for telegram ID %s: %v", telegramID, err)
	}
	return user, err
}

func insertOrUpdateUser(user Player) error {
	log.Printf("Inserting or updating user for telegram ID: %s", user.TelegramID)
	collection := db.Collection("users")
	filter := bson.M{"telegramId": user.TelegramID}
	update := bson.M{"$set": user}
	_, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Error inserting/updating user for telegram ID %s: %v", user.TelegramID, err)
	}
	return err
}