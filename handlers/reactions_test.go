package handlers

import (
	"blue/config"
	"blue/models"
	"testing"
)

func TestAdminReactionChanceDependsOnTalkingToBot(t *testing.T) {
	cfg := &config.Config{AdminUserIDs: map[int64]bool{8: true}}
	msg := &models.Message{
		From: &models.User{ID: 8},
		Chat: &models.Chat{ID: -1, Type: "supergroup"},
		Text: "nice",
		ReplyToMessage: &models.Message{
			From: &models.User{ID: 99, IsBot: true},
		},
	}

	if got := reactionChanceDenominator(msg, 99, "maple_bot"); got != 2 {
		t.Fatalf("expected 1/2 chance for a reply to the bot, got 1/%d", got)
	}
	msg.ReplyToMessage = nil
	if got := reactionChanceDenominator(msg, 99, "maple_bot"); got != 20 {
		t.Fatalf("expected 1/20 chance for an ordinary message, got 1/%d", got)
	}

	msg.Text = "yo @maple_bot!"
	if got := reactionChanceDenominator(msg, 99, "maple_bot"); got != 2 {
		t.Fatalf("expected exact mention to use 1/2 chance, got 1/%d", got)
	}
	msg.Text = "yo @maple_bot_fake"
	if got := reactionChanceDenominator(msg, 99, "maple_bot"); got != 20 {
		t.Fatalf("longer username must use 1/20 chance, got 1/%d", got)
	}

	msg.From.ID = 9
	if eligibleForAdminReaction(msg, cfg) {
		t.Fatal("non-admin message was eligible for a reaction")
	}
}

func TestAdminReactionEligibilityIncludesPrivateChats(t *testing.T) {
	cfg := &config.Config{AdminUserIDs: map[int64]bool{8: true}}
	msg := &models.Message{
		From: &models.User{ID: 8},
		Chat: &models.Chat{ID: 8, Type: "private"},
		Text: "hello",
	}
	if !eligibleForAdminReaction(msg, cfg) {
		t.Fatal("admin DM should be eligible for a reaction")
	}
}

func TestChannelPostIsEligibleWithoutExposingAuthor(t *testing.T) {
	cfg := &config.Config{AdminUserIDs: map[int64]bool{8: true}}
	msg := &models.Message{
		Chat: &models.Chat{ID: -100, Type: "channel"},
		Text: "new post",
	}
	if !eligibleForAdminReaction(msg, cfg) {
		t.Fatal("channel post should be treated as admin-authored")
	}

	handler := &Handler{customReactionEmojis: []string{"approved"}}
	if got := handler.customReactionEmoji(msg); got != "approved" {
		t.Fatalf("expected approved channel reaction, got %q", got)
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
