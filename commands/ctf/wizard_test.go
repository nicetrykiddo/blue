package ctf

import (
	"testing"
	"time"
)

func TestWizardSessionsStayScopedToUserAndTopic(t *testing.T) {
	first := wizardKey{chatID: 1, threadID: 10, userID: 100}
	otherTopic := wizardKey{chatID: 1, threadID: 11, userID: 100}
	otherUser := wizardKey{chatID: 1, threadID: 10, userID: 101}
	setWizard(first, wizardState{kind: "new", stage: 2})
	t.Cleanup(func() { deleteWizard(first) })

	if state, ok := getWizard(first); !ok || state.stage != 2 {
		t.Fatal("expected the exact user/topic session")
	}
	if _, ok := getWizard(otherTopic); ok {
		t.Fatal("session leaked into another topic")
	}
	if _, ok := getWizard(otherUser); ok {
		t.Fatal("session leaked to another user")
	}
}

func TestSetWizardRemovesExpiredSessions(t *testing.T) {
	expired := wizardKey{chatID: 2, threadID: 20, userID: 200}
	current := wizardKey{chatID: 2, threadID: 20, userID: 201}
	setWizard(expired, wizardState{expires: time.Now().Add(-time.Minute)})
	setWizard(current, wizardState{expires: time.Now().Add(time.Minute)})
	t.Cleanup(func() {
		deleteWizard(expired)
		deleteWizard(current)
	})

	if _, ok := getWizard(expired); ok {
		t.Fatal("expired wizard session was not removed")
	}
	if _, ok := getWizard(current); !ok {
		t.Fatal("current wizard session was removed")
	}
}
