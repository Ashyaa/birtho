package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ashyaa/birtho/bot"
	"github.com/ashyaa/birtho/log"
	"github.com/sirupsen/logrus"
)

var (
	logger logrus.Logger
	b      *bot.Bot
)

func init() {
	var err error
	logger = log.New()
	b, err = bot.New(&logger)
	if err != nil {
		panic(err)
	}
}

func main() {
	// Wait here until CTRL-C or other term signal is received.
	logger.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	b.Stop()
}
