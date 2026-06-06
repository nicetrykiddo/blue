package main

import (
	"blue/api"
	"blue/commands/ctf"
	"blue/config"
	"blue/database"
	"blue/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	bot := api.NewBot(cfg.BotToken)

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	handler := handlers.NewHandler(bot, db, cfg)
	ctf.StartDailyScheduler(bot, db, cfg)

	if err := bot.SetWebhookWithOptions(api.WebhookOptions{
		URL:             cfg.WebhookURL,
		SecretToken:     cfg.WebhookSecret,
		CertificateFile: cfg.WebhookCertificate,
	}); err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}
	log.Println("Webhook registered")

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		update, err := api.ParseWebhookUpdate(r, cfg.WebhookSecret)
		if err != nil {
			log.Printf("Error parsing webhook: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go handler.HandleUpdate(*update)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		if cfg.GroupID != 0 {
			bot.SendMessage(cfg.GroupID, "```\ngoing offline\n```")
		}
		os.Exit(0)
	}()

	if cfg.GroupID != 0 {
		bot.SendMessage(cfg.GroupID, "```\nback online\nmode: webhook\n```")
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Bot started. Listening on %s...", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
