package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ashyaa/birtho/bot"
	"github.com/ashyaa/birtho/log"
)

func main() {
	logger := log.New()
	err := bot.New(logger)
	if err != nil {
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	logger.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
