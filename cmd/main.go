package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	gr "github.com/tryy3/minatsugo/github-release"
)

func init() {
	godotenv.Load()
}

func main() {
	announcementManager := gr.NewAnnouncementManager()

	defer announcementManager.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
