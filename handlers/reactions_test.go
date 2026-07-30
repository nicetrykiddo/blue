package handlers

import (
	"blue/config"
	"blue/models"
	"testing"
)

func TestAdminReactionChanceDependsOnTalkingToBot(t *testing.T) {
	cfg := &config.Config{AdminUserIDs: map[int64]bool{8: true}}
	msg := &models.Message{
		MessageID: 16,
		From:      &models.User{ID: 8},
		Chat:      &models.Chat{ID: -1, Type: "supergroup"},
		Text:      "nice",
		ReplyToMessage: &models.Message{
			From: &models.User{ID: 99, IsBot: true},
		},
	}

	if !shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("expected eligible reply to the bot to receive a reaction")
	}
	msg.MessageID++
	if shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("expected half of replies to skip reactions")
	}

	msg.ReplyToMessage = nil
	msg.MessageID = 28
	if !shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("expected low-chance ordinary message to receive a reaction")
	}
	msg.MessageID++
	if shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("expected most ordinary messages to skip reactions")
	}

	msg.MessageID = 16
	msg.Text = "yo @maple_bot!"
	if !shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("expected exact bot mention to use the higher chance")
	}
	msg.Text = "yo @maple_bot_fake"
	if shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("longer username must not count as a bot mention")
	}

	msg.From.ID = 9
	if shouldReactToAdmin(msg, cfg, 99, "maple_bot") {
		t.Fatal("non-admin message received a reaction")
	}
}

func TestCustomReactionIDsComeOnlyFromRepliedMessage(t *testing.T) {
	msg := &models.Message{
		Entities: []models.MessageEntity{{Type: "custom_emoji", CustomEmojiID: "ignored"}},
		ReplyToMessage: &models.Message{Entities: []models.MessageEntity{
			{Type: "custom_emoji", CustomEmojiID: "123"},
			{Type: "custom_emoji", CustomEmojiID: "123"},
		}},
	}

	ids := repliedCustomEmojiIDs(msg)
	if len(ids) != 1 || ids[0] != "123" {
		t.Fatalf("expected one replied custom emoji, got %#v", ids)
	}
}
