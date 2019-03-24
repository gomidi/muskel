

# Einbindung externer Dateien

Nutzt man MuSkel am Computer, insbesondere in Verbindung mit dem muskel Programm (siehe Kapitel 3),
so macht es Sinn, die Inhalte eines Stücks auf mehrere Dateien zu verteilen und mit dem include Befehl
in die Hauptdatei einzubinden. Es können sowohl im Kopfbereich, als auch im Notationsbereich weitere 
MuSkel-Dateien eingebunden werden. Im Kopfbereich werden dabei nur die Templates importiert (siehe Kapitel 5) 
und alle Eigenschaften etc. ignoriert. Im Notationsbereich ist es wichtig, dass in der eingebundenen Datei 
nur Spuren vorliegen, die auch in der Hauptdatei, die einbindet, vorhanden sind.
Die eingebundenen Dateien müssen entweder im gleichen Ordner wie die Hauptdatei liegen, oder im Konfigurationsverzeichnis des muskel-Programms.


Beispiel:
Inhalt der Hauptdatei main.mskl

```
=
    | piano    | flute        |
$include("verse1")
$include("chorus")
```

Inhalt von verse1.mskl

```
=
    | piano    | flute        | 
4/4@100
1   | c        |              |
2   | d        | e            |
3   |          | f            |
```

Inhalt von chorus.mskl

```
=
    | piano    | flute        | 
1   | f        |              |
2   | f        | d            |
3   |          | d            |
```

Ergebnis:

```
=
    | piano    | flute        | 
4/4@100
1   | c        |              |
2   | d        | e            |
3   |          | f            |

1   | f        |              |
2   | f        | d            |
3   |          | d            |
```

Im Konfigurationsverzeichnis von muskel ist eine Datei vorinstalliert, die Templates für die gängigen GM-kompatiblen Schlagzeugsound beinhaltet und mit 

$include("drumnotes")

in den Kopfbereich eingebunden werden können:

kd  kick drum
sn  snare drum
ho  open hihat
hc  closed hihat
rd  ride cymbal
tb  tambourine
tl  low tom
tm  middle tom
th  high tom
sh  crash cymbal


# Das muskel Hilfsprogramm

Das muskel Hilfsprogramm versteht das muskel-Sketch Dateiformat, welches die MuSkel-Syntax beinhaltet
mit dem Zeichenvorrat von ASCII oder UTF-8 (ohne Byte-Order-Marker) und einem Unix-Zeilenumbruch.

Es bietet die folgende Unterstützung:

- automatisches Formatieren mit tabellarischer Spaltenausrichtung
- Einbindung von weiteren Dateien
- Export in das Standard MIDI File-Format (Multitrack) bei voller Syntaxunterstützung
- "ausrollen" von Takt- und Bereichswiederholungen, Templates und Zufällen zur Analyse
- Prüfung der Syntax und präzise Fehlermeldungen bei Syntaxfehlern
- automatisches Beobachten von Änderungen mit Anschließender Formatierung/Export
- Feedback beim Beobachten mit Popupfenstern
- Löschen, Umbenennen und Synchronisieren von Spuren auch in eingebundenen Dateien
- Umbenennen von Templates in Definition und Benutzung, auch in eingebundenen Dateien
- automatisches Abspielen der exportierten SMF-Datei auf manchen Plattformen (erfordert GM-kompatiblen Player)
- lauffähig unter Linux, Android, MS Windows, MacosX, `*BSD`, 32bit und 64bit

## Installation

Für folgende Betriebssystem werden Binärdateien bereitgestellt:

- MS Windows 64bit
- Linux 64bit

Die Binärdateien können unter 

https://gitlab.com/gomidi/muskel

heruntergeladen werden

Für alle anderen Systeme ist es nötig, Go version 1.12 oder höher zu installieren und dann folgenden
Befehl auszuführen:

```
go install gitlab.com/gomidi/muskel/cmd/muskel
```

## Nutzung

Der Befehl 

```
muskel help
```

listet die verschiedenen Optionen und Funktionen auf.

# Editoren

Unter https://gitlab.com/gomidi/muskel/editors finden sich
Syntaxdateien und andere Unterstützung für folgende Editoren:

- vim
- sublime text
- micro
- notepad++

# DAWs

Unter https://gitlab.com/gomidi/muskel/daws/reaper.md findet sich
eine Anleitung zur Verwendung mit Reaper.
