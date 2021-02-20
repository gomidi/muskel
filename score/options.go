package score

import (
	"gitlab.com/gomidi/muskel/items"
)

type Option func(s *Score)

func Column(colname string) Option {
	return func(s *Score) {
		s.mainCol = colname
	}
}

func Debug() Option {
	return func(s *Score) {
		//s.mainCol = colname
		items.DEBUG = true
	}
}

func Sketch(sketchname string) Option {
	return func(s *Score) {
		s.mainSketch = sketchname
	}
}

func PrintBarComments() Option {
	return func(s *Score) {
		s.printBarComments = true
	}
}

func NoEmptyLines() Option {
	return func(s *Score) {
		s.noEmptyLines = true
	}
}

func CutOut() Option {
	return func(s *Score) {
		s.cutout = true
	}
}

func SoloGroup(groupID uint) Option {
	return func(s *Score) {
		s.soloGroup = groupID
	}
}

func CSV(seperator rune) Option {
	return func(s *Score) {
		s.csvSeperator = seperator
	}
}
