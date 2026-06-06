package ctf

import (
	"blue/models"
	"fmt"
	"time"
)

const (
	voiceLoadingLive       = "<i>Checking what is currently on fire...</i>"
	voiceLoadingUpcoming   = "<i>Reading the calendar like it owes us money...</i>"
	voiceLoadingLiveDone   = "<i>Found the smoke. Sorting it by how soon everyone suffers...</i>"
	voiceLoadingFutureDone = "<i>Calendar decoded. Removing the boring bits...</i>"
	voiceStorageMissing    = "CTF storage is not plugged in. The calendar is standing here with empty hands."
	voiceLiveError         = "Live CTF fetch tripped. CTFtime blinked first; try again in a bit."
	voiceUpcomingError     = "Upcoming CTF fetch tripped. The schedule drawer jammed."
	voiceNoPermission      = "Nice try, but this button is behind the staff door."
)

func voiceDailyList(now time.Time) listViewCopy {
	return listViewCopy{
		Title:   "CTF watchlist",
		Subline: fmt.Sprintf("%s | next %d day(s) | freshly scraped, lightly judged", localStamp(now), defaultLookaheadDays()),
		Empty:   "Nothing worth waking the scoreboard for in this window.",
		Footer:  "Hit <b>I'm in</b> if you are actually showing up. Enough hands go up, a dedicated topic appears.",
	}
}

func voiceLiveList(now time.Time) listViewCopy {
	return listViewCopy{
		Title:   "Live CTFs",
		Subline: fmt.Sprintf("%s | these are already moving", localStamp(now)),
		Empty:   "Nothing live right now. The scoreboard is pretending to be peaceful.",
	}
}

func voiceUpcomingList(now time.Time, days int) listViewCopy {
	return listViewCopy{
		Title:   "Upcoming CTFs",
		Subline: fmt.Sprintf("%s | next %d day(s) | calendar ambushes, sorted", localStamp(now), days),
		Empty:   "Clean window. Suspicious, but clean.",
	}
}

func voiceManualCreated(id int, title string) string {
	return fmt.Sprintf("CTF <b>#%d</b> is on the board: %s", id, safe(title))
}

func voiceSavedError() string {
	return "Could not save that CTF. The database rejected the offering."
}

func voiceNotFound() string {
	return "I looked. That CTF is not in the drawer."
}

func voiceNoTopic(id int) string {
	return fmt.Sprintf("That CTF has no war room yet. Run <code>/ctftopic %d</code> first.", id)
}

func voiceNoTopicShort() string {
	return "That CTF has no war room yet."
}

func voiceEditFailed() string {
	return "Telegram refused the edit. Annoying, but honest."
}

func voiceRefreshFailed() string {
	return "Telegram refused the refresh. It does that sometimes."
}

func voiceTopicCreateFailed() string {
	return "Could not open the war room. Give the bot topic powers and I will try again."
}

func voiceDigestFailed() string {
	return "Could not push the digest. The pipe made a face."
}

func voiceEdited(id int) string {
	return fmt.Sprintf("Updated <b>#%d</b>. The opening note has new orders.", id)
}

func voiceRefreshed(id int) string {
	return fmt.Sprintf("Refreshed <b>#%d</b>. The topic message is back in fighting shape.", id)
}

func voiceTopicCreated(id int, title string) string {
	return fmt.Sprintf("War room opened for <b>#%d</b>: %s", id, safe(title))
}

func voiceTopicExists(id int) string {
	return fmt.Sprintf("<b>#%d</b> already has a war room. No duplicate doors today.", id)
}

func voiceDigestSent() string {
	return "Digest shipped to the CTF topic. Calendar violence delivered."
}

func voiceUsageEdit() string {
	return "Shape it like this: <code>/ctfedit &lt;ctf_id&gt; &lt;new initial topic message&gt;</code>"
}

func voiceUsageRefresh() string {
	return "Point me at one: <code>/ctfrefresh &lt;ctf_id&gt;</code>"
}

func voiceUsageTopic() string {
	return "Point me at one: <code>/ctftopic &lt;ctf_id&gt;</code>"
}

func voiceCallbackBooting() string {
	return "CTF desk is still booting"
}

func voiceCallbackNoUser() string {
	return "I cannot tell who clicked that"
}

func voiceCallbackBadButton() string {
	return "That button came in sideways"
}

func voiceCallbackMissingCTF() string {
	return "That CTF fell off the map"
}

func voiceCallbackVoteFailed() string {
	return "Could not save the vote. Rude database moment."
}

func voiceCallbackJoined(createdTopic bool, joined bool, count int) string {
	switch {
	case createdTopic:
		return fmt.Sprintf("You're in. War room opened with %d on deck.", count)
	case joined:
		return fmt.Sprintf("You're in. Roster is at %d.", count)
	default:
		return fmt.Sprintf("Already counted. Still %d on deck.", count)
	}
}

func voiceRosterJoined(user *models.User) string {
	return fmt.Sprintf("%s is on the roster.", userMention(user))
}

func localStamp(t time.Time) string {
	return t.In(appLocation()).Format("Mon, 02 Jan 2006 15:04 MST")
}
