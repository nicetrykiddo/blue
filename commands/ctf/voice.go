package ctf

import (
	"blue/models"
	"fmt"
	"time"
)

const (
	voiceLoadingLive       = "<i>checking what is on fire rn...</i>"
	voiceLoadingUpcoming   = "<i>checking the calendar...</i>"
	voiceLoadingLiveDone   = "<i>found some, sorting...</i>"
	voiceLoadingFutureDone = "<i>calendar done, sorting...</i>"
	voiceStorageMissing    = "ctf desk is still waking up, poke me again in a sec"
	voiceLiveError         = "live board blinked. try again in a bit yk"
	voiceUpcomingError     = "calendar drawer jammed. classic. try again in a bit"
	voiceNoPermission      = "nice try but staff only yk"
)

func voiceDailyList(now time.Time) listViewCopy {
	return listViewCopy{
		Title:   "ctf watchlist",
		Subline: fmt.Sprintf("%s | next %d day(s)", localStamp(now), defaultLookaheadDays()),
		Empty:   "nothing in this window rn",
		Footer:  "hit <b>I'm in</b> if you're playing. enough votes and i open a topic.",
	}
}

func voiceLiveList(now time.Time) listViewCopy {
	return listViewCopy{
		Title:   "live ctfs",
		Subline: fmt.Sprintf("%s | live rn", localStamp(now)),
		Empty:   "nothing live rn",
	}
}

func voiceUpcomingList(now time.Time, days int) listViewCopy {
	return listViewCopy{
		Title:   "upcoming ctfs",
		Subline: fmt.Sprintf("%s | next %d day(s)", localStamp(now), days),
		Empty:   "nothing coming up rn",
	}
}

func voiceManualCreated(title string) string {
	return fmt.Sprintf("ctf added: <b>%s</b>", safe(title))
}

func voiceSavedError() string {
	return "that ctf did not stick. try again"
}

func voiceNotFound() string {
	return "i checked, that ctf is not in the drawer"
}

func voiceNoTopic(id int) string {
	return fmt.Sprintf("that ctf has no topic yet. run <code>/ctftopic %d</code> first", id)
}

func voiceNoTopicShort() string {
	return "that ctf has no topic yet"
}

func voiceEditFailed() string {
	return "telegram said no to editing it"
}

func voiceRefreshFailed() string {
	return "telegram said no to refreshing it"
}

func voiceTopicCreateFailed() string {
	return "topic did not open. check my perms"
}

func voiceDigestFailed() string {
	return "digest did not send. try again"
}

func voiceEdited(title string) string {
	return fmt.Sprintf("updated <b>%s</b>. opening msg changed.", safe(title))
}

func voiceRefreshed(title string) string {
	return fmt.Sprintf("refreshed <b>%s</b>. topic msg is back in shape.", safe(title))
}

func voiceTopicCreated(title string) string {
	return fmt.Sprintf("topic opened for <b>%s</b>", safe(title))
}

func voiceTopicExists(title string) string {
	return fmt.Sprintf("<b>%s</b> already has a topic.", safe(title))
}

func voiceDigestSent() string {
	return "digest sent to the ctf topic."
}

func voiceImOutNotTopic() string {
	return "use /imout inside the ctf topic"
}

func voiceImOutNoUser() string {
	return "idk who typed that"
}

func voiceImOutNotFound() string {
	return "this topic is not linked to a ctf"
}

func voiceImOutWasNotIn() string {
	return "you were not in this roster"
}

func voiceImOutDone(count int) string {
	return fmt.Sprintf("removed you. roster is at %d.", count)
}

func voiceImOutClosed() string {
	return "no one left. closing topic."
}

func voiceImOutFailed() string {
	return "could not remove you rn. try again"
}

func voiceUsageEdit() string {
	return "shape it like this: <code>/ctfedit &lt;ctf_id&gt; &lt;new topic msg&gt;</code>"
}

func voiceUsageRefresh() string {
	return "give me the id bruv: <code>/ctfrefresh &lt;ctf_id&gt;</code>"
}

func voiceUsageTopic() string {
	return "give me the id bruv: <code>/ctftopic &lt;ctf_id&gt;</code>"
}

func voiceCallbackBooting() string {
	return "ctf desk still booting"
}

func voiceCallbackNoUser() string {
	return "idk who clicked that"
}

func voiceCallbackBadButton() string {
	return "that button is weird"
}

func voiceCallbackMissingCTF() string {
	return "that ctf fell off the map"
}

func voiceCallbackVoteFailed() string {
	return "vote did not stick. poke it again"
}

func voiceCallbackJoined(createdTopic bool, joined bool, count int) string {
	switch {
	case createdTopic && joined:
		return fmt.Sprintf("you are in. topic opened with %d on deck.", count)
	case createdTopic:
		return fmt.Sprintf("topic opened. roster is at %d.", count)
	case joined:
		return fmt.Sprintf("you are in. roster is at %d.", count)
	default:
		return fmt.Sprintf("already counted. still %d on deck.", count)
	}
}

func voiceRosterJoined(user *models.User) string {
	return fmt.Sprintf("%s joined the roster.", userMention(user))
}

func localStamp(t time.Time) string {
	return t.In(appLocation()).Format("Mon, 02 Jan 2006 15:04 MST")
}
