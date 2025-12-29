package main

import (
	"blue/api"
	"blue/cache"
	"blue/config"
	"blue/handlers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	bot := api.NewBot(cfg.BotToken)

	stickerCache := cache.NewStickerCache(bot)
	if err := stickerCache.Load("HANG_SEED_MINI3"); err != nil {
		log.Printf("Warning: Failed to load sticker set: %v", err)
	}

	handler := handlers.NewHandler(bot, stickerCache)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		if cfg.GroupID != 0 {
			bot.SendMessage(cfg.GroupID, "<code>[-] System Offline\n| Status: Shutdown\n| Mode: Maintenance</code>")
		}
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	log.Println("Bot started. Polling for updates...")

	if cfg.GroupID != 0 {
		bot.SendMessage(cfg.GroupID, "<code>[+] System Online\n| Status: Active\n| Mode: Production</code>")
	}

	offset := 0
	for {
		updates, err := bot.GetUpdates(offset, 30)
		if err != nil {
			log.Printf("Error getting updates: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			go handler.HandleUpdate(update)
		}

		time.Sleep(1 * time.Second)
	}
}
