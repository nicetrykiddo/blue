package commands

import "testing"

func TestPublicMenuDoesNotExposeAdminCommands(t *testing.T) {
	adminOnly := map[string]bool{
		"newctf": true, "editctf": true, "openctf": true, "refreshctf": true,
		"ctfsync": true, "allowreaction": true, "removereaction": true, "clearreactions": true,
	}
	for _, command := range PublicMenu() {
		if adminOnly[command.Command] {
			t.Fatalf("admin command %q leaked into public menu", command.Command)
		}
	}
}
