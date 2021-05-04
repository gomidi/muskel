module gitlab.com/gomidi/muskel/cmd/muskel

go 1.14

require (
	github.com/metakeule/observe v0.0.5
	gitlab.com/gomidi/lilypond v0.0.5 // indirect
	gitlab.com/gomidi/midi v1.23.0
	gitlab.com/gomidi/midicatdrv v0.1.0
	gitlab.com/gomidi/muskel v1.0.0
	gitlab.com/gomidi/muskel/csv v1.2.4
	gitlab.com/gomidi/muskel/xlsx v1.2.4
	gitlab.com/gomidi/smfimage v0.14.7 // indirect
	gitlab.com/metakeule/config v1.18.9
	golang.org/x/sys v0.0.0-20201211090839-8ad439b19e0f
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
)

replace (
	gitlab.com/gomidi/muskel => ../../
	gitlab.com/gomidi/muskel/csv => ../../csv
	gitlab.com/gomidi/muskel/xlsx => ../../xlsx
)
