package main

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messageMonitor() {
	defer g.wg.Done()
	logInfo.Println("Starting message monitor")
	defer logWarn.Println("Stopping message monitor")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates, err := g.bot.GetUpdatesChan(u)
	if err != nil {
		logErr.Printf("Update failed: %v\n", err)
	}

outer:
	for {
		select {
		case <-g.shutdown:
			break outer
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			if update.Message.IsCommand() {
				handleMessage(update.Message)
			}
		}
	}
}

func commandIsForMe(t string) bool {
	command := strings.SplitN(t, " ", 2)[0] // Return first substring before space, this is entire command

	i := strings.Index(command, "@") // Position of @ in command
	if i == -1 {                     // Not in command
		return true // Assume command is for everybody, including this bot
	}

	return strings.ToLower(command[i+1:]) == strings.ToLower(g.bot.Self.UserName)
}

func handleMessage(m *tgbotapi.Message) {
	if !commandIsForMe(m.Text) {
		return
	}

	logInfo.Printf("%s", m.Text)
	simpleSend := func(s string) { g.bot.Send(tgbotapi.NewMessage(m.Chat.ID, s)) }

	switch strings.ToLower(m.Command()) {
	case "id":
		simpleSend(fmt.Sprintf("Hi, %s %s, your Telegram user ID is %d", m.From.FirstName, m.From.LastName, m.From.ID))
	case "info":
		simpleSend(fmt.Sprintf("This chat's ID is %d", m.Chat.ID))
	case "start", "help":
		handleHelp(m)
	case "hi":
		simpleSend("Hi!")
	case "ping":
		simpleSend("Pong!")
	case "pong":
		simpleSend("Ping!")
	case "mug", "auw", "dood", "splash", "auwundo", "mugundo", "reset":
		handleMug(m)
	}
}

func handleHelp(m *tgbotapi.Message) {
	msg := "This bot warns you at special times. Add a time at which you want to be warned every day using '/add'"
	g.bot.Send(tgbotapi.NewMessage(m.Chat.ID, msg))
}

func handleMug(m *tgbotapi.Message) {
	var format = "Muggen %d - Wij %d"
	var muggen, wij int
	fmt.Sscanf(m.Chat.Title, format, &muggen, &wij)
	logInfo.Printf("Muggen = %d", muggen)
	logInfo.Printf("Wij = %d", wij)

	switch strings.ToLower(m.Command()) {
	case "mug", "dood", "splash":
		wij++
	case "auw":
		muggen++
	case "auwundo":
		wij--
	case "mugundo":
		muggen--
	case "reset":
		muggen = 0
		wij = 0
	}
	var newTitle = fmt.Sprintf(format, muggen, wij)

	resp, err := g.bot.SetChatTitle(tgbotapi.SetChatTitleConfig{
		ChatID: m.Chat.ID,
		Title:  newTitle,
	})
	if err != nil {
		logErr.Printf("%s\n", err)
	}
	if !resp.Ok {
		logErr.Printf("%s\n", resp.Result)
		logErr.Printf("%s\n", resp.Description)
	}
}
