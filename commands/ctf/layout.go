package ctf

import (
	"blue/database"
	"fmt"
	"strings"
)

type responseView struct {
	builder strings.Builder
}

func newResponseView(title, subline string) *responseView {
	view := &responseView{}
	view.line("<b>%s</b>", safe(title))
	if subline != "" {
		view.line("<i>%s</i>", safe(subline))
	}
	view.blank()
	return view
}

func (v *responseView) line(format string, args ...interface{}) {
	v.builder.WriteString(fmt.Sprintf(format, args...))
	v.builder.WriteString("\n")
}

func (v *responseView) rawLine(text string) {
	v.builder.WriteString(text)
	v.builder.WriteString("\n")
}

func (v *responseView) blank() {
	v.builder.WriteString("\n")
}

func (v *responseView) field(label, value string) {
	v.line("<b>%s:</b> %s", safe(label), value)
}

func (v *responseView) text() string {
	return trimTelegramMessage(v.builder.String())
}

type listViewCopy struct {
	Title   string
	Subline string
	Empty   string
	Footer  string
}

func renderEventList(copy listViewCopy, events []database.CTFEvent, limit int) string {
	view := newResponseView(copy.Title, copy.Subline)
	if len(events) == 0 {
		view.rawLine(safe(copy.Empty))
		return view.text()
	}

	limit = min(len(events), limit)
	for i := 0; i < limit; i++ {
		view.rawLine(formatEventCard(i+1, events[i]))
		if i < limit-1 {
			view.blank()
		}
	}

	if copy.Footer != "" {
		view.blank()
		view.rawLine(copy.Footer)
	}

	return view.text()
}
