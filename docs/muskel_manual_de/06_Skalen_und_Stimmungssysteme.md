# Skalendefinitionen

Um Stimmungen zu erzeugen, sind am Ende des Tages Pitchbend-Messages notwendig.

Technisch gesehen das einfachste wäre es, man definierte für jeden Ton einen fixen Offset, der anhand des Pitchbend-Range
umgerechnet wird in einen PitchbendOffset, der den neuen virtuellen "Nullpunkt" des Pitchbend-Rades markiert und sowohl
bei absoluten, Skalennoten, als auch normalen Pitchbend-Messages zum Einsatz kommt.

Auf eine Funktion/ein Interface heruntergebrochen wäre das

type Tuner interface {
	Tune(key uint8, pitchbend int16, pitchbendRange uint8) (tunedKey uint8, tunedPitchbend int16)
}

dieses interface ist von jedem Tuning zu erfüllen.

Die naheliegendste und wahrscheinlich kompatibelste Variante einen Pitchshift neutral zum pitchbendrange 
zu definieren, ist die Einheit cent, da wir hierbei uns auf eine Verschiebung zum gleichschwebenden Stimmungssystem 
beziehen.

Wir wollen verschiedene Tunings und Skalen parallel spielen können.

Für Skalen können wir pro Instrument eine Zeile "Scale" haben, die die allgemeingültige Skala überschreibt.
Dann können wir im Score mit vorangestellem Backslash vor eine absolute Note, den Basiston der Skala
in der Zeit verändern

\a"

setzt den Basiston auf a". Diese Setzung gilt erst für alle nachfolgenden Items.

Für Tunings haben wir pro Instrument eine Zeile "Tuning" und global für den Takt die vorangestellte Raute.

#pythagorean

setzt das Tuning global auf das pythagoräische.

Um eigene Skalen zu definieren verwenden wir wieder den Backslash.
Hier werden nun absolute Noten verwendet, wobei a" den virtuellen Nullpunkt markiert und die halbton-abstände 
über den abstand daraus berechnet werden. z.B.

\my_scale:                                       d",e",f#",g",a",c"

resultiert in (halbton-differenz zum Nullpunkt)  -7,-5,-3 ,-2, 0,3

Die absoluten Noten werden aber intern auch gespeichert.
Beim verwenden einer Skala, wird der neue Nullpunkt anhand einer absoluten Note gesetzt. Hierbei wird zunächst
die "nächste" absolute Note gemäß der Definition gesucht, gemerkt und die entsprechende halbtondifferenz gemerkt.
Bei der Suche der nächsten Note wird wie folgt vorgegangen:
Der eingehängte Ton wird solange oktaviert, bis die Differenz zu einer der bei der Definition der Skala
gesetzten absoluten Note minimal ist. Angestrebt werden 0-3 Halbtöne. Dann wird dieser "nächste" Ton und die
Differenz dazu gemerkt.
Wird nun ein Skalenton angefordert, So wird diese halbton-differenz auf den Skalenschritt aufaddiert und dann 
wird der "nächste" Ton verwendet, um anhand dessen Differenz die Schritte auf der Skala zu gehen.

Beispiel:

nehmen wir die obere Skala und hängen sie auf b' ein.
Hierdurch wird als nächster Ton c" gewählt mit einer Differenz von +13 halbtönen.
Wenn wir jetzt den Skalenton ^-3 anfordern (was ja einer Differenz von -3 gegenüber dem eingehängten Nullpunkt entspricht)
so wählen wir den "nächsten" Ton c" als Nullpunkt, gehen drei Schritte runter (landen beim f#") und addieren
+13 halbtöne (landen beim g"' als Zielton).

Eine Skala wird in der Oktave wiederholt, wenn das Ende der Skala erreicht ist.

also in der oben angegebenen Skala ist der nächste Ton nach c" der Ton d"' und der Ton unter d" der Ton c'
Diese Variante erlaubt es, auch Skalen zu definieren, die partiell nicht oktavrein sind, z.B.

  d#', f',g',a#', c#', d",e",f#",g",a",c"

Hier wäre der Ton oberhalb von c" der Ton d#"' und der Ton unterhalb von d#' wäre c. Jedoch besteht zwischen c#' und c"
keine Oktavreinheit, wie auch zwischen a#' und a" etc.
auch vollständige nicht oktavreine Scalen können definiert werden, wenn die relevanten 5-7 Oktaven komplett absolut definiert werden.

Die Form der Definition ermöglicht, oktavreine Skalen mit minimalem aufwand zu definieren, sie ermöglicht aber auch
die definition von Skalen mit absoluten Ankerpunkten, die nicht oder nur partiell Oktavrein sind.
Die Transposition beim einhängen erfolgt so immer als minimale Verschiebung gegenüber absoluten Ankerpunkten, so dass
der Charakter (etwa bei Instrumenten) maximal gewahrt wird.

Für die Definition eines Tunings können wir nun ähnlich vorgehen.

Hier fangen wir mit der Raute vor dem Namen des Tunings an und haben eine Zuordnung von absoluten Noten zu entweder cent-Differenzen 
zur entsprechenden Frequenzen der absoluten Note, oder direkt in absoluten Frequenzwerten. Die cent-Werte werden hinter dem Doppelpunkt angegeben, Frequenzen hinter dem =:

#my_tune:  d":23,d#":12,e":-12,f#":40,g":30,a"=400,c":10

Angegebene Frequenzen werden direkt in cent differenzen() umgerechnet und intern werden nur die cent Differenzen gespeichert.
Die Auswahl der entsprechenden Cent-Differenz einer absoluten Note erfolgt nach dem gleichen Prinzip des "nächsten Tones" und der
Oktavierungen (oder eben keiner), wie bei den Skalendefinitionen. Hier wird als Cent-Differenz eines absoluten Tones immer
die Cent-Differenz des "nächsten Tones" (sei er auch oktaviert) verwendet, wobei als Nähe die Anzahl der Halbtonschritte gemäß
Benennung genommen wird. Wenn man alle Halbtöne in einem Tuning definiert, hat man also die größte Kontrolle.
 

Es wäre noch zu Überlegen,ob es für Tuning und Skalendefinition nicht nett wäre, wenn die Viertel-Tönige und Achteltönige 
Notation für die absolten Noten Verwendung finden würde. Damit mehrere kleinabstände innerhalb einer Skala möglich wären:

c    = ganzton c
c^   = achtelton zwischen c und c°
c°   = viertelton zwischen c und c#
c°^  = achtelton zwischen c° und c#
c#   = halbton c#
c#^  = achtelton zwischen c# und c#°
c#°  = viertelton zwischen c# und d
c#°^ = achtelton zwischen c#° und d
d    = ganzton d



# Definition eines Stimmungssystems