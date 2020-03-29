module gitlab.com/gomidi/muskel/cmd/muskel

go 1.14

require (
	github.com/gen2brain/beeep v0.0.0-20180718162406-4e430518395f
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/metakeule/observe v0.0.0-20170406094839-eba094fd212a
	gitlab.com/gomidi/midi v1.14.1
	gitlab.com/gomidi/muskel v1.0.0
	gitlab.com/metakeule/config v1.13.0
	golang.org/x/sys v0.0.0-20190209173611-3b5209105503 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
)

replace gitlab.com/gomidi/muskel => ../../
