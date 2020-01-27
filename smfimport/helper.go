package smfimport

import "gitlab.com/gomidi/midi"

type positionedMsg struct {
	absPos uint64
	msg    midi.Message
}

type colsKey struct {
	trackNo int16
	channel uint8
}

type sortPositionedMsg []positionedMsg

func (s sortPositionedMsg) Len() int {
	return len(s)
}

func (s sortPositionedMsg) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortPositionedMsg) Less(a, b int) bool {
	return s[a].absPos < s[b].absPos
}
