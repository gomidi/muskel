package stops_handling

import "gitlab.com/gomidi/midi/v2/smf"

type Event struct {
	pos32th     uint
	absPosition uint
	msg         smf.Message
	trackNoteOn bool
}

// sort them by absPosition
type Events []Event

func (s Events) Less(a, b int) bool {

	if s[a].absPosition == s[b].absPosition {
		am := s[a].msg.IsMeta()
		bm := s[b].msg.IsMeta()

		if am && !bm {
			return true
		}

		return false
	}

	return s[a].absPosition < s[b].absPosition
}

func (s Events) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s Events) Len() int {
	return len(s)
}
