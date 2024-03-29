package main

import (
	"encoding/gob"
	"github.com/BurntSushi/toml"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	logErr  = log.New(os.Stderr, "[ERRO] ", log.Ldate+log.Ltime+log.Ltime+log.Llongfile)
	logWarn = log.New(os.Stdout, "[WARN] ", log.Ldate+log.Ltime)
	logInfo = log.New(os.Stdout, "[INFO] ", log.Ldate+log.Ltime)
	g       = global{shutdown: make(chan bool)}
)

func main() {
	/////////////
	// STARTUP
	//////////////

	// Parse settings file
	_, err := toml.DecodeFile("settings.toml", &g.c)
	if err != nil {
		logErr.Println(err)
		return
	}

	// Create new bot
	g.bot, err = tgbotapi.NewBotAPI(g.c.Apikey)
	if err != nil {
		logErr.Println(err)
	}

	logInfo.Printf("Running as @%s", g.bot.Self.UserName)

	// Create waitgroup, for synchronized shutdown
	var wg sync.WaitGroup
	g.wg = &wg

	// All messages are received by the messageMonitor
	wg.Add(1)
	go messageMonitor()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)

	logInfo.Println("All routines have been started, awaiting kill signal")

	///////////////
	// SHUTDOWN
	///////////////

	// Program will hang here
	select {
	case <-sigs:
		close(g.shutdown)
	case <-g.shutdown:
	}
	println()
	logInfo.Println("Shutdown signal received. Waiting for goroutines")

	// Shutdown after all goroutines have exited
	g.wg.Wait()
	logWarn.Println("Shutting down")
}

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}
func Load(path string, o interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		dec := gob.NewDecoder(file)
		err = dec.Decode(o)
	}
	file.Close()
	return err
}
