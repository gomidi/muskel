module gitlab.com/gomidi/muskel/cmd/muskel

go 1.14

require (
	github.com/metakeule/observe v0.0.5
	gitlab.com/gomidi/lilypond v0.0.5
	gitlab.com/gomidi/midi v1.23.7
	gitlab.com/gomidi/midicatdrv v0.3.7
	gitlab.com/gomidi/muskel v1.0.0
	gitlab.com/gomidi/muskel/xlsx v1.2.4
	gitlab.com/metakeule/config v1.21.0
)

replace (
	gitlab.com/gomidi/muskel => ../../
	gitlab.com/gomidi/muskel/csv => ../../csv
	gitlab.com/gomidi/muskel/xlsx => ../../xlsx
)
