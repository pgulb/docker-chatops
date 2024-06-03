package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
	"github.com/joho/godotenv"
	"github.com/pgulb/docker-chatops/docker"
)

var allowedChatIds []int64
var logsReplyKeyboard *reply.ReplyKeyboard
var restartReplyKeyboard *reply.ReplyKeyboard

func message(text string, b *bot.Bot, ctx context.Context, chatId int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   text,
	})
}

func messageAll(text string, b *bot.Bot, ctx context.Context) {
	for _, chatId := range allowedChatIds {
		message(text, b, ctx, chatId)
	}
}

func logMessage(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			log.Printf("%d say: %s", update.Message.From.ID, update.Message.Text)
		}
		next(ctx, b, update)
	}
}

func loadDotenv() string {
	err := godotenv.Load("/app/.env")
	if err != nil {
		log.Fatal(err)
	}
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is empty")
	}
	allowedChatIdsCommas := os.Getenv("ALLOWED_CHAT_IDS")
	if allowedChatIdsCommas == "" {
		log.Fatal("ALLOWED_CHAT_IDS is empty")
	}
	allowedChatIdsStr := strings.Split(allowedChatIdsCommas, ",")
	for _, chatIdStr := range allowedChatIdsStr {
		chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		allowedChatIds = append(allowedChatIds, chatId)
	}
	return token
}

func main() {
	token := loadDotenv()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	opts := []bot.Option{
		bot.WithMiddlewares(logMessage),
	}
	b, err := bot.New(token, opts...)
	if nil != err {
		log.Fatal(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/ps", bot.MatchTypeExact, ps)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/logs", bot.MatchTypeExact, logs)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/restart", bot.MatchTypeExact, restart)

	log.Println("*** Chatops bot started ***")
	messageAll("*Chatops bot started*", b, ctx)
	b.Start(ctx)
}

func ps(ctx context.Context, b *bot.Bot, update *models.Update) {
	resp, err := docker.ListContainers(ctx)
	if err != nil {
		log.Println(err.Error())
		message(err.Error(), b, ctx, update.Message.Chat.ID)
	} else {
		message(resp, b, ctx, update.Message.Chat.ID)
	}
}

func initLogKeyboard(b *bot.Bot, ctx context.Context) error {
	logsReplyKeyboard = reply.New(
		b,
		reply.WithPrefix("logs_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)
	ctr, err := docker.ListContainersNamesOnly(ctx)
	if err != nil {
		return err
	}
	for _, name := range ctr {
		logsReplyKeyboard.Button(fmt.Sprintf("Logs %v", name),
			b, bot.MatchTypeExact, onReplyLogs)
		logsReplyKeyboard.Row()
	}
	logsReplyKeyboard.Button("Cancel Logs", b, bot.MatchTypeExact, onReplyLogs)
	return nil
}

func logs(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := initLogKeyboard(b, ctx)
	if err != nil {
		log.Println(err.Error())
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Select container:",
		ReplyMarkup: logsReplyKeyboard,
	})
}

func onReplyLogs(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Text == "Cancel Logs" {
		message("Cancelled.", b, ctx, update.Message.Chat.ID)
		return
	}
	if strings.HasPrefix(update.Message.Text, "Logs ") {
		resp, err := docker.TailLogs(ctx, strings.Split(update.Message.Text, " ")[1])
		if err != nil {
			log.Println(err.Error())
			message(err.Error(), b, ctx, update.Message.Chat.ID)
		} else {
			message(resp, b, ctx, update.Message.Chat.ID)
		}
	}
}

func initRestartKeyboard(b *bot.Bot, ctx context.Context) error {
	restartReplyKeyboard = reply.New(
		b,
		reply.WithPrefix("restart_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)
	ctr, err := docker.ListContainersNamesOnly(ctx)
	if err != nil {
		return err
	}
	for _, name := range ctr {
		restartReplyKeyboard.Button(fmt.Sprintf("Restart %v", name),
			b, bot.MatchTypeExact, onReplyRestart)
		restartReplyKeyboard.Row()
	}
	restartReplyKeyboard.Button("Cancel Restart", b, bot.MatchTypeExact, onReplyRestart)
	return nil
}

func restart(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := initRestartKeyboard(b, ctx)
	if err != nil {
		log.Println(err.Error())
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Select container:",
		ReplyMarkup: restartReplyKeyboard,
	})
}

func onReplyRestart(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Text == "Cancel Restart" {
		message("Cancelled.", b, ctx, update.Message.Chat.ID)
		return
	}
	if strings.HasPrefix(update.Message.Text, "Restart ") {
		resp, err := docker.RestartContainer(ctx, strings.Split(update.Message.Text, " ")[1])
		if err != nil {
			log.Println(err.Error())
			message(err.Error(), b, ctx, update.Message.Chat.ID)
		} else {
			message(resp, b, ctx, update.Message.Chat.ID)
		}
	}
}
