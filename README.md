# Telegram Bot

A simple Telegram bot built with Go using only the Telegram Bot API (no third-party packages).

## Structure

- `main.go` - Entry point and polling loop
- `config/` - Configuration management
- `api/` - Telegram API wrapper
- `models/` - Data structures for Telegram types
- `handlers/` - Message and command handlers

## Setup

1. Get a bot token from [@BotFather](https://t.me/botfather) on Telegram
2. Set the environment variable:
   ```bash
   $env:TELEGRAM_BOT_TOKEN="your_token_here"
   ```
3. Run the bot:
   ```bash
   go run main.go
   ```

## Adding New Features

- Add new commands in `handlers/handlers.go`
- Add new API methods in `api/telegram.go`
- Add new types in `models/types.go`
