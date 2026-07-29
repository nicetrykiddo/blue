package ctf

import (
	"blue/models"
	"testing"
)

func TestSelectTopicIconUsesTitleCategory(t *testing.T) {
	stickers := []models.Sticker{
		{Emoji: "🏆", CustomEmojiID: "trophy"},
		{Emoji: "🌐", CustomEmojiID: "web"},
	}
	if got := selectTopicIcon("Friday Web CTF", stickers); got != "web" {
		t.Fatalf("expected web icon, got %q", got)
	}
	if got := selectTopicIcon("Friday CTF", stickers); got != "trophy" {
		t.Fatalf("expected trophy fallback, got %q", got)
	}
}
