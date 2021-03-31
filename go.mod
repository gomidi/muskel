module gitlab.com/gomidi/muskel

go 1.14

require (
	github.com/emersion/go-appdir v1.1.2
	gitlab.com/gomidi/midi v1.23.0
	gitlab.com/metakeule/config v1.18.8
	gitlab.com/gomidi/muskel/csv v1.2.4
	gitlab.com/gomidi/muskel/xlsx v1.2.4
	gitlab.com/gomidi/quantizer v0.0.1
)

replace (
	gitlab.com/gomidi/muskel/csv => ./csv
	gitlab.com/gomidi/muskel/xlsx => ./xlsx
)
