# MIDI

MIDI ist ein Standard in der digitalen Musikerzeugung seit 1980.
MuSkel unterstützt MIDI 1.0 durch das Programm muskel und den gelieferten SMF-Export.
SMF ist die Abkürzung für Standard MIDI Fileformat und beinhaltet die Spezifikation,
wie MIDI Nachrichten in Dateien gespeichert werden können.


## SMF-Export

Hier wird besprochen, wie das muskel Programm, muskel-Sketch-Dateien in das SMF-Format exportiert.
Andere Programme können bei diesem Weg andere Strategien verfolgen.

Es wird vor allem das Format 1 verwendet (Multitrack), welches eine beliebige Anzahl von
Spuren unterstützt, obwohl die Gesamtzahl der MIDI-Kanäle weiterhin 16 ist.

Allerdings können durch geschicktes Importieren in diverse Programme falls gewünscht
16 unabhängige MIDI Kanäle pro Spur genutzt werden.

Das muskel Programm wandelt jede MuSkel Spur in einen Track einer SMF-Datei um, wobei immer zwei zusätzliche
Tracks geschrieben werden: Der erste Track beinhaltet die Taktwechsel und Bereichsmarker (Bereich-Benennungen
werden in Marker und Cuepoints konvertiert) und der zweite Track beinhaltet die Tempowechsel.

Die weiteren Tracks beinhalten jeweils eine Spur des MuSkel-Datei, wobei die Reihenfolge der Spuren die gleiche
ist, wie die innerhalb der MuSkel-Datei. Darüberhinaus werden diese Tracks benannt mit dem gleichen Namen der
jeweiligen Spur, wobei etwaige umgebende spitze Klammern des Namens entfernt werden.
Da verschiedene Programme die Tracknamen aus unterschiedlichen MIDI meta messages extrahieren, werden nach dem Gießkannenprinzip folgende Meta Messages auf den Namen gesetzt:
ProgramName, TrackSequenceName, InstrumentName

## Spureigenschaften

Es gibt einige unterstützte Spureigenschaften, welche direkt im entsprechenden MIDI-Track in entsprechende
MIDI Messages umgewandelt werden:

Ch   setzt den MIDI Kanal. Das muskel Programm erwartet nur einen Kanal pro Spur, obwohl mehrere Spuren den
     gleichen Kanal haben können. Alle MIDI Channel Messages, wie Noten, Pitchbend, Controller etc. liegen
     auf diesem Kanal. Ist kein Kanal gesetzt, wird die Spur nicht in die SMF-Datei exportiert
Prog setzt eine einmalige ProgramChange message, um ein Programm mit der entsprechenden Nummer im Klangerzeuger
     zu wählen
Bank setzt eine einmalige BankSelect Controller message, um eine Bank mit der entsprechenden Nummer im Klangerzeuger zu wählen
Volume setzt eine einmalige VolumeMSB Controller message, um die Laustärke des MIDI Kanals im Klangerzeuger zu 
setzen
PbRange setzt eine PitchBendSensitivityRPN Controller message, um den Wertebereich der Pitchbend Messages zu wählen (siehe unter Glissando, wird leider nicht von allen Klangerzeugern unterstützt), setzt aber auch die interne Formel
zur Umrechnung von Halbtonschritten in Pitchbend-Werte
Trans Transponiert normale und Skalennoren (aber keine reinen MIDINotes. s.u.) um die entsprechende Zahl an Halbtonschritten
FileGroup ermöglicht beim Export das gleichzeitige exportieren in verschiedene Dateien (s.u.)

## Velocitywerte

Dynamikwerte werden in Velocitywerte von NoteOn messages umgewandelt. Hierbei gilt folgende Umrechnung:

++++  123
+++   108
++     93
+      78
*      63
-      48
--     33
---    18
----    5

auf diesen Wert wird für jede Note eine Zufallszahl zwischen 0 und 4 aufaddiert, um das Spiel etwas
weniger statisch zu gestalten.

## Dateigruppen

Exportiert man in eine SMF-Datei mit einem Dateinamen, der einen Platzhalter %s enthält, so wird dieser
beim Export ersetzt mit der FileGroup-Eigenschaft der jeweiligen Spur, so dass alle Spuren mit der gleichen FileGroup-Eigenschaft in der gleichen SMF-Datei landen. Jede SMF-Datei hat jedoch die gleichen zwei ersten Spuren
(Taktwechsel und Tempowechsel).

## MIDI-Controller

Die folgenden Ereignisse sind reserviert für dezidierte MIDI-Messages

n steht für eine normale note, wie z.B. c", die in einen MIDI-Key umgerechnet wird
v steht für einen Wert 0-127
c steht für einen controller 0-127
pb steht für einen Pitchbend-Wert -8192 bis 8192

AT(v)     kanalweite Aftertouch Message mit Wert v
PB(pb)    kanalweite Pitchbend Message mit Wert pb
PT(n,v)   Tasten-/Notenspezifische Aftertouch Message für Note n mit Wert v
CC(c,v)   ContinouseController Message für Controller c mit Wert v

Durch Anfügen von einer oder zwei Tilden ~ können auch hier linear bzw. exponentielle Verläufe zum
nächstfolgenden Wert definiert werden. Es ist jedoch angeraten, hierbei verschiedene Spuren für die jeweiligen
Messagetypen bzw. Controller anzulegen (die dann in die gleiche Datei exportiert werden und den gleichen Kanal bekommen). Das macht es sowohl leserlicher, als auch den Export zuverlässiger.

## Glissandi

Glissandi werden in Pitchbend Messages umgerechnet. Grundlage hierfür ist der pro Spur angegebene PitchbendRange
PbRange, der angibt wieviele Halbtöne die Spanne zwischen höchstem (8192) und niedrigstem (-8192) Pitchbendwert
umfasst. Der Standardwert sind 2 Halbtöne. Es gibt eine PitchBendSensitivityRPN mit der manche Klangerzeuger es
erlauben, den Pitchbendwert von außen zu setzen. Diese message wird von muskel auch in den entsprechenden Track der SMF-Datei geschrieben, aber nicht alle Klangerzeuger reagieren darauf. Manche haben gar einen festen Pitchbendrange. In diesem Fall hilft nur, durch ausprobieren (z.B. mittels Pitchbendcontroller am Keyboard) den
Pitchbendrange per Gehör festzustellen und dann die richtige Anzahl an Halbtönen in der PbRange Eigenschaft der
MuSkel Spur einzutragen. 
 

## Zufallsnoten

Wahrscheinlichkeiten und Zufallsauswahl wird beim SMF-Export per Computer gewürfelt und entsprechend verwendet.

## MIDINote

Es gibt noch einen anderen Weg, MIDI note on messages zu erzeugen, als mit normalen Noten oder Skalennoten:
Das MN Ereignis:

MNk    erzeugt eine MIDINote des keys k (0-127)

Auch hier kann man Dynamikwerte oder Glissandozeichen oder Stakkatozeichen anhängen. Diese MIDI noten werden aber
weder beim Transponieren mittel Trans Eigenschaft der Spur, noch durch Skalen transponiert und sind somit
für noten/keys geeignet, die eine feste Belegung beim Klangerzeuger haben, z.B. Keyswitches oder die Sounds
eines Drumsamplers.
