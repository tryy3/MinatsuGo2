package main

import (
	"fmt"
	"log"
	"net/http"
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

	// Quick and dirty ping webserver
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Pong")
	})

	fmt.Println("Run baby ruuun")
	log.Fatalf("Error starting webserver: %v", http.ListenAndServe(":80", nil))

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
