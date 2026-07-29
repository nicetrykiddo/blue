package ctf

import (
	"blue/config"
	"blue/models"
	"testing"
)

func TestCTFManagementDefaultsToDenied(t *testing.T) {
	previous := cfg
	t.Cleanup(func() { cfg = previous })

	user := &models.User{ID: 42}
	cfg = nil
	if canManageCTFs(user) {
		t.Fatal("nil config must not grant CTF management")
	}

	cfg = &config.Config{AdminUserIDs: map[int64]bool{}}
	if canManageCTFs(user) {
		t.Fatal("empty admin list must not grant CTF management")
	}

	cfg.AdminUserIDs[user.ID] = true
	if !canManageCTFs(user) {
		t.Fatal("configured admin should be allowed")
	}
}

func TestHTTPURLRejectsOtherSchemes(t *testing.T) {
	if got := buttonURL("ftp://example.com/file"); got != "" {
		t.Fatalf("expected ftp URL to be rejected, got %q", got)
	}
	if got := buttonURL("https://example.com/event"); got != "https://example.com/event" {
		t.Fatalf("expected HTTPS URL, got %q", got)
	}
}
