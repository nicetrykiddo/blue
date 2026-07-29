package handlers

import (
	"blue/config"
	"blue/models"
	"testing"
)

func TestAdminReactionIsSparseAndAdminOnly(t *testing.T) {
	cfg := &config.Config{AdminUserIDs: map[int64]bool{8: true}}
	msg := &models.Message{
		MessageID: 16,
		From:      &models.User{ID: 8},
		Chat:      &models.Chat{ID: -1, Type: "supergroup"},
		Text:      "nice",
	}

	if !shouldReactToAdmin(msg, cfg) {
		t.Fatal("expected matching admin message to receive a reaction")
	}
	msg.MessageID++
	if shouldReactToAdmin(msg, cfg) {
		t.Fatal("expected most admin messages to receive no reaction")
	}
	msg.From.ID = 9
	if shouldReactToAdmin(msg, cfg) {
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
