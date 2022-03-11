package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ashyaa/birtho/bot"
)

func main() {
	bot.Run()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
