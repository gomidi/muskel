# MuSkel: Musical Sketch Language

## Einführung

`MuSkel` ist eine *Domain Specific Language* (`DSL`), d.h. eine Sprache, die für einen bestimmten 
Einsatzzweck optimiert ist. Der Einsatzzweck ist in diesem Falle das Notieren und Skizzieren 
musikalischer Ideen und Kompositionen.

Es wurde besonderen Wert darauf gelegt, das `MuSkel` nur minimale technische Voraussetzungen erfordert 
und so neutral wie möglich ist.

Daher kann man diese Spezifikation auch als eine Art *Stenografie für Musik* benutzen - mit Stift und Papier.

Schon bei dieser Verwendung bietet sie gegenüber etwa der `westlichen Standardnotation` diverse Vorteile, z.B. 
- die Unabhängigkeit von Notenpapier (keine Notenlinien)
- weniger Schreib- und Lesefehleranfälligkeit
- die Möglichkeit, Templates und Abkürzungen zu verwenden
- das einfachere Erlernen der Notation von Noten
- die Möglichkeit, klangliche Veränderungen von Synthesizern zu Notieren
- die Möglichkeit, relativ zu einer Skala zu notieren
- die Möglichkeit, diverse Stimmungssysteme und Skalen parallel zu nutzen
- die Möglichkeit, eigene Stimmungssysteme und Skalen zu definieren

`MuSkel` lässt sich aber auch an praktisch jedem (auch älteren) Computersystem schreiben, da die 
verwendeten Zeichen alle Teil des `ASCII Standards` sind.

Die einzige Voraussetzung ist ein Editor, der Zeilenumbrüche im Unix-Standard (`LF` bzw. `\n`) darstellen 
kann. Dies ermöglichen praktisch alle Editoren, die zum Programmieren verwendet werden und auch viele 
einfachere Editoren.

Da die Syntax nicht besonders komplex ist, ist es leicht, Hervorhebungsdateien für bestimmte Editoren zu 
erstellen, um das Schreiben weiter zu vereinfachen. Für den Editor `Vim`, der auf unzählige Betriebssysteme 
portiert wurde, ist bereits eine Hervorhebung und weitere Unterstützung implementiert.

Darüber hinaus gibt es das Programm `muskel`, welches unter `MS Windows`, `MacOSX`, `Linux` und den 
`BSD-Systemen` genutzt werden kann und diverse Werkzeuge bietet. 

Hierbei sind besonders hervorzuheben, 

1. die automatische Formatierung, welche das Schreiben in Verbindung
mit einem Editor, der geänderte Dateien automatisch neu laden kann sehr erleichtert 
(z.B. `Vim`, `Sublime Text` und viele andere).

2. der automatische Export in das *Standard-MIDI-Dateiformat* (`SMF`), welches von praktisch jedem 
Musikprogramm (Sequencerprogram, Digital Audio Workstation, DAW) unterstützt wird und in diesen Programmen 
leicht eine Weiterverarbeitung und Weiterentwicklung der Kompositionen ermöglicht. Hierbei ist besonders 
die DAW `Reaper` hervorzuheben, die es ermöglicht, dass die exportierten `SMF`-Dateien mit einem 
Tastenkürzel neu geladen werden, was einen schnellen Workflow ermöglicht. 

3. Besondere Funktionen, um etwa Spuren und Templates an verschiedenen Stellen und Dateien gleichzeitig umzubenennen, etc.

In Kombination mit einem Notationsprogramm, wie etwa `MuseScore`, lassen sich die `SMF`-Dateien in dieses 
importieren und damit leicht in die `westliche Standardnotation` konvertieren.

Zusammenfassend lässt sich die `MuSkel`-Sprache folgendermaßen charakterisieren:
- "analog" und digital nutzbar
- erweiterte Möglichkeiten gegenüber der `westlichen Standardnotation`
- mobil und technikneutral nutzbar
- in digitaler Form leicht mit gängiger Musiksoftware weiterzuverarbeiten
- leicht erlernbar

Da man nicht die gesamte Syntax beherrschen muss, um `MuSkel` zu nutzen, kann sich der Anfänger auf die 
fundamentalen Aspekte konzentrieren, die sehr leicht zu erlernen sind und die fortgeschrittenen 
Möglichkeiten erlernen, sobald er sie benötigt.

Um den Einstieg so leicht wie möglich zu machen, werden zunächst die Prinzipien der Syntax beschrieben 
ohne, dass wir uns mit dem Programm `muskel` oder dem Thema `MIDI` befassen.

Diese Prinzipen können wir mit Papier und Stift, oder mit einem Editor unserer Wahl ausprobieren und lernen.

Das Kapitel `01_Grundlagen` erklärt diese Grundlagen und sollte in jedem Falle komplett und der Reihe 
nach durchgearbeitet werden. Es ist erfreulicherweise weder lang noch schwierig und ermöglicht bereits 
eine Vielzahl von Anwendungen. Die Inhalte und Begriffe des Grundlagenkapitels werden jedoch an allen 
weiteren Stellen vorausgesetzt.

Wird `MuSkel` mit dem Computer genutzt (was sehr zu empfehlen ist), so sollten die Kapitel `03_Hinweise_und_Hilfsmittel_zur_Nutzung_am_Computer` und `04_MIDI_und_SMF-Export` durchgearbeitet werden. 
Hier finden sich Hinweise und Installationen, die einmalig durchgeführt werden und dann bei konsequenter 
Nutzung das Schreiben, Ändern, Hören und Weiterverarbeiten von Musik mit `MuSkel` erheblich beschleunigen 
und vereinfachen.

**Alle weiteren Kapitel sind optional und können nach Bedarf und auch Ausschnittweise gelesen werden**.

Das Kapiel `02_Vertiefung` bearbeitet seltener benötigte Eigenschaften einer Musiknotation und deckt in 
Verbindung mit den Grundlagen die kompletten Möglichkeiten der `westlichen Standardnotation ab` (und geht
selbst darüber hinaus). Den Abschluss dieses Kapitels bildet eine Übersicht der Unterschiede zur 
`westlichen Standardnotation`.

Die fortgeschrittenen Themen der Kapitel 5 und 6 sollten erst bearbeitet werden, wenn eine gewisse 
Routine in der Verwendung von `MuSkel` besteht, da sie den Einsteiger überfordern können. 

Die Templates im Kapitel 5 helfen bei der Komposition mit `MuSkel` - insbesondere dem Aufbau von 
motivischen und rhythmischen Verwandtschaften und Zusammenhängen.

Skalen- und Stimmungssysteme im Kapitel 6 setzen Kenntnisse der Skalen und Stimmungsproblematik voraus 
und wenden sich an fortgeschrittene Nutzer, die mit eigenen Skalen und Stimmungssystemen experimentieren 
oder historische bzw. außereuropäische Skalen und Stimmungen verwenden wollen.

 