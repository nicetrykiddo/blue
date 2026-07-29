package main

import (
	"blue/api"
	"blue/commands"
	"blue/commands/ctf"
	"blue/config"
	"blue/database"
	"blue/handlers"
	"context"
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

	if err := bot.SetCommands(commands.Menu()); err != nil {
		log.Printf("Could not update Telegram command menu: %v", err)
	}

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

	if cfg.GroupID != 0 {
		bot.SendHTMLMessage(cfg.GroupID, "<pre>back online\nmode: webhook</pre>")
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownSignal, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	go func() {
		<-shutdownSignal.Done()
		log.Println("Received shutdown signal...")
		if cfg.GroupID != 0 {
			bot.SendHTMLMessage(cfg.GroupID, "<pre>going offline</pre>")
		}
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("Bot started. Listening on %s...", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
