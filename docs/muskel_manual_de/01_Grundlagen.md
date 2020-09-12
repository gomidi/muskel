# Grundlagen

## Aufbau eines MuSkel Sketches

Die `MuSkel`-Sprache erlaubt es, eine oder mehrere musikalische "Notizen" (engl. sketches) anzufertigen
und zu einem Musikstück zu kombinieren.
Im einfachsten Fall besteht das Stück aus einem einzigen `sketch`, dem `=SCORE`.
Mit Papier und Stift umfasst er eine oder mehrere fortlaufende (nummerierte) Seiten, in der digitalen 
Welt steht er in einer Dateien, deren Namen auf `.mskl` endet und in der Regel `main.mskl` heisst.

Ein `sketch`ist eine Tabelle, die einen Namen hat, der mit dem Gleichheitszeichen beginnt (z.B. `=SCORE`)
und deren Spalten die Stimmen oder Spuren (engl. tracks) beinhalten. Die erste Zeile der Tabelle beinhaltet
den Namen der Tabelle und die Namen der Spalten/Spuren.

Die Zeitachse schreitet von oben nach unten fort und jede weitere Zeile stellt entweder einen Taktwechsel 
oder einen bestimmten Zeitpunkt im Takt dar.

Die erste Spalte hat eine besondere Funktion - alle weiteren Spalten sind Spuren, in denen 
`Ereignisse` stattfinden können.

Ein `Ereignis` wäre z.B., das eine Note beginnt. **Alle Ereignisse
in einer Zeile finden gleichzeitig statt.**

Die erste Spalte ist speziell. Sie die *Position im laufenden Takt* bzw. *Eigenschaften des laufenden Taktes*.

Schauen wir uns den einfachsten möglichen `sketch` an:

```
=SCORE | piano | 
#
   1   |       |
```

Zunächst haben wir den Namen der Tabelle `=SCORE`. Durch das Gleichheitszeichen am Anfang des Namens wird
festgelegt, dass es sich um einen `sketch` handelt. Dann folgt durch einen senkrechten Balken abgetrennt der
Name der ersten Spur "piano". Ein weiterer senkrechter Balken schließt die Spalte.

In der Regel entspricht eine `Spur` einem Instrument bzw. einer `Stimme` eines Instrumentes. Hier wäre das Instrument wahrscheinlich ein Klavier, aber letztlich ist die Benennung der Spur frei wählbar.

In der zweiten Zeile steht eine Raute, die den Start eines Taktes markiert. Hier werden keine senkrechten
Balken verwendet, da der Takt für alle Spuren gilt.

Es folgt in der ersten Spalte der dritten Zeile eine Taktposition; hier der *erste Schlag* (Beat/Viertelnote) im Takt, 
also der Start des Taktes (wie gezählt wird), ausgedrückt durch die `1`. 

Doch wie lang dauert der Takt? Wenn nichts anderes definiert wird, ist der erste Takt ein `4/4`-Takt und 
sein Tempo liegt dann bei 120 Schlägen pro Minute (`BPM`).

Dann folgen die Ereignisse, die zu diesem Zeitpunkt in den verschiedenen Spuren stattfinden.
Hier haben wir nur die Spur "piano" und es findet an dieser Taktposition kein Ereignis statt. Daher folgen
lediglich zwei senkrechte Balken, um die Spalte zu begrenzen.

Dieser `MuSkel Sketch` besteht aus einem Takt und einer Spur, auf der im Takt nichts passiert.

In der digitalen Welt, müssen wir noch etwas ergänzen, wenn wir die Datei in eine SMF-Datei (Standard-MIDI-Format)
umwandeln wollen, um sie mit anderen Musikprogrammen weiterverarbeiten zu können.

Dann brauchen wir eine weitere Tabelle, die `TRACK`heisst und allgemeine Informationen zu den Spuren beinhaltet,
die für MIDI relevant sind. Die wichtigste Eigenschaft ist dabei der MIDI-Kanal, auf dem die Ereignisse der
Spur gesendet werden sollen.

Hierbei lernen wir noch eine wichtige Eigenschaft von Tabellen:
Es gibt keine leere Zeilen innerhalb einer Tabelle. Stattdessen werden leere Zeilen verwendet, um Tabellen
zu beenden, bzw. abzugrenzen.

vollständig sieht unsere `main.mskl` Datei dann wie folgt aus:

```
TRACK   | piano | 
Channel | 1     |

=SCORE  | piano | 
#
   1    |       |
```

Hier haben wir der "piano"-Spur in der `TRACK`Tabelle den MIDI-Kanal 1 zugewiesen. Der MIDI-Standard erlaubt
16 Kanäle. Welchen Sie für welche Spur verwenden, ist Ihnen überlassen. Sie können die Ereignisse eines
Kanals aber auf mehrere Spuren verteilen, indem Sie den Spuren den gleichen Kanal zuweisen.

Im Folgenden lassen wir wegen der Lesbarkeit die erforderliche `TRACK`Tabelle weg, es sei denn, sie wird gerade
besprochen.

## Noten und Takte

So ohne `Ereignisse` ist es noch nicht wirklich spannend, also lassen wir etwas passieren:

```
=SCORE |piano| 
#
   1   |c    |
   2   |d    |
   3   |e    |
   4   |f    |
#
   1   |g    |
```

Hier haben wir nun die Noten `c`,`d`,`e` und `f` auf den Schlägen `1`,`2`,`3` und `4` des ersten Takts und die Note `g` auf dem 1. Schlag des zweiten Taktes.

Wenn zum Erzeugen des neuen Taktes lediglich die Raute `#` verwendet wird, werden die Eigenschaften des
neuen Taktes vom vorigen Takt übernommen.

Wollen wir jedoch die Taktart ändern, so verwenden wir die "Bruchnotation" mit Querstrich `/`:

```
=SCORE   |piano| 
#
       1 |c    |
       2 |d    |
       3 |e    |
       4 |f    |
# 3/4
       1 |g    |
```

Hier wechselt der Takt zum `3/4`-Takt, d.h. der erste Takt ist ein `4/4`-Takt (Standard) und der Zweite ein `3/4`-Takt. Um die Lesbarkeit zu erhöhen, wurden die Bezeichnung der Schlagpositionen eingerückt. Leerzeichen bzw. freie Flächen können in MuSkel frei verteilt werden.

Das bedeutet auch, dass die Spalten "schief" sein können:

```
=SCORE  |piano| 
#
        1  |c    |
    2   |d    |
    3  |e    |
      4    |f    |
# 3/4
       1   |g    |
```

Dies ist zwar nicht besonders leserlich, aber valide und kann, wenn wenig Zeit vorhanden ist, hilfreich sein. Verwendet man `MuSkel` am Computer, so kann man mit dem Programm `muskel` die Datei beobachten lassen, so dass das Programm die Spalten automatisch sauber anordnet, wenn man Änderungen speichert. In Verbindung mit einem Editor, der automatisch geänderte Dateien neu laden kann (z.B. `Vim`, `Sublime Text`), kann man somit sehr komfortabel arbeiten, da man sich nicht um die Spaltenanordnung kümmern muss.

## Dauer der Töne und Monophonie einer Spur

Wir erkennen nun die Startpunkte der Töne, aber wie definieren wir die Dauer einer Note? Eine Besonderheit von `MuSkel` ist, dass alle Spuren zunächst *einstimmig* (monophon) sind (zur Ausnahme kommen wir später), d.h. es erklingt pro `Spur` immer nur ein Ton zur gleichen Zeit und dieser endet, wenn der nächste Ton erklingt. Also so ähnlich, wie bei der menschlichen Stimme, oder bei Blasinstrumenten.

Wenn wir wollen, dass ein Ton endet, ohne dass ein neuer Ton erklingt, so drücken wir dies durch das `Stille`-Zeichen `*` aus:

```
=SCORE  |piano| 
#
      1 |c    |
      2 |*    |
      3 |e    |
      4 |f    |
# 3/4
      1 |g    |
```

Hier endet die Note `c` auf dem zweiten Schlag (dauert also ein `Viertel`-Note lang).

Zeilen im Takt, in denen keine `Ereignisse` stehen, können weggelassen werden, also würde hier

```
=SCORE   |piano| 
#
       1 |c    |
       3 |e    |
       4 |f    |
# 3/4
       1 |g    |
```

das `c` erst enden, wenn auf dem 3.Schlag das `e` folgt (das `c` dauert also eine `Halbe`-Note lang).

Wollen wir, dass schon der erste Takt ein `3/4`-Takt wird, so schreiben wir:

```
=SCORE   |piano| 
# 3/4
       1 |c    |
       3 |e    |
#
       1 |g    |
```

## Noten und Oktavlage

Nun gibt es einige `c` in verschiedenen Oktaven. Das beschriebene `c` ist das sogenannte "kleine c" (MIDI-Note ...)
Aufwärts kann man durch Verwenden von Hochkommata oktavieren. Es ergeben sich
- kleines c: `c`
- eingestrichenes c: `c'`  
- zweigestrichenes c: `c"`  
- dreigestrichenes c: `c"'` 

Analog gilt dies für alle anderen Noten.

Großgeschrieben, ergibt sich das "große c" und mit nachgestellten Hochkommata die unteren Oktaven:
- großes c: `C`   
- Oktave unter dem großen c: `C'`  
- zwei Oktaven unter dem großen c: `C"`  
- drei Oktaven unter dem großen c: `C"'` 

`Halbtöne` werden mit nachgestellter Raute `#` bezeichnet. Die Raute steht nach dem Buchstaben aber vor den Hochkommata:
- kleines cis: `c#`  
- eingestrichenes cis: `c#'` 
usw.

Man kann stattdessen auch eine Note um einen Halbton erniedrigen durch ein nachgestelltes b:
- kleines es: `eb`  
- eingestrichenes es: `eb'` 
usw.

Die `Standardstimmung` ist die `gleichschwebende Stimmung`. Daher gilt `cis`= `des` usw. Abweichende Stimmungen können aber definiert werden.

## Zeitpositionen im Takt

Für die Zeitpositionen im Takt gibt es auch noch feinere Aufteilungen, als die Schläge (Viertel) im Takt. Teilen wir einen Schlag (Viertel), so bekommen wir zwei Achtel. Diese werden mit "und" gezählt und mit `&` geschrieben:

```
=SCORE  | piano | 
#
    1   | c     |
    1&  | c'    |
    2   | d     |
    2&  | c'    |
    3   | e     |
    3&  | c'    |
    4   | f     |
    4&  | c'    |
```

Hier haben wir ein `c'` auf jedem "und" bzw `&` (auch genannt Offbeat). So verkürzen sich alle Notenlängen im Beispiel auf `Achtel`-Noten. Es gibt noch feinere Aufteilungen, wie 16tel Noten:

```
=SCORE  |piano| 
#
    1   |c    |
    1,  |d'   |
    1&  |c'   |
    1&, |d'   |
    2   |d    |
    2,  |d'   |
    2&  |c'   |
    2&, |d'   |
    3   |e    |
    3,  |d'   |
    3&  |c'   |
    3&, |d'   |
    4   |f    |
    4,  |d'   |
    4&  |c'   |
    4&, |d'   |
```


oder 32tel Noten:

```
=SCORE  |piano| 
#
    1   |c    |
    1;  |e'   |
    1,  |d'   |
    1,; |e'   |
    1&  |c'   |
    1&; |e'   |
    1&, |d'   |
    1&,;|e'   |
    2   |d    |
    2;  |e'   |
    2,  |d'   |
    2,; |e'   |
    2&  |c'   |
    2&; |e'   |
    2&, |d'   |
    2&,;|e'   |
```

Alternativ kann man die Zeitpositionen auch als Dezimalzahlen schreiben bezogen auf Viertel (in englischer Schreibweise mit dem Punkt `.` als Dezimaltrenner). Dabei entspricht `1.5` dann `1&` und `1.75` entspricht `1&,` usw. 
Damit lässt sich der obere `sketch`auch so schreiben:

```
=SCORE    |piano| 
#
    1     |c    |
    1.125 |e'   |
    1.25  |d'   |
    1.375 |e'   |
    1.5   |c'   |
    1.625 |e'   |
    1.75  |d'   |
    1.875 |e'   |
    2     |d    |
    2.125 |e'   |
    2.25  |d'   |
    2.375 |e'   |
    2.5   |c'   |
    2.625 |e'   |
    2.75  |d'   |
    2.875 |e'   |
```


## Stakkato

Will man kurze Stakkato-Noten schreiben, kann es schnell unübersichtlich werden, z.B.:

```
=SCORE    |piano|
# 
     1    |c    |
     1;   |*    |
     1,   |d'   |
     1,;  |*    |
     1&   |c'   |
     1&;  |*    |
     1&,  |d'   |
     1&,; |*    |
     2    |d    |
     2;   |*    |
     2,   |d'   |
     2,;  |*    |
     2&   |c'   |
     2&;  |*    |
     2&,  |d'   |
     2&,; |*    |
```

Dafür gibt es eine spezielle Notation, die die Noten auf 64tel verkürzt und damit Stakkato-Feeling erzeugt:

```
=SCORE  |piano | 
#
    1   |c     |
    1.  |d'::: |
    1&  |c'    |
    1&. |d'::: |
    2   |d     |
    2.  |d'::: |
    2&  |c'    |
    2&. |d'::: |
```

Im obigen Beispiel wäre jedes `d'` eine 64tel, da die Note mit drei Doppelpunkten `:::` endet.
Bei zwei Doppelpunkten `::` wäre es eine 32tel und bei einem Doppelpunkt `:` eine 16tel.

## Überbindungen

Eine Note dauert bis die nächste Note oder eine Pause (Stille) notiert ist, **auch über Taktwechsel hinweg**:

```
=SCORE  |piano| 
#
    1   |c    |
#
    1&  |d    |
```

Hier endet das `c` erst auf 1& des folgenden Taktes.

Wollen wir, dass es schon auf dem ersten Schlag des folgenden Taktes endet, so müssen wir dort eine Pause setzen:

```
=SCORE  |piano| 
#
    1   |c    |
#
    1   |*    |
    1&  |d    |
```

Wollen wir die letzte Note wiederholen, so verwenden wir das isolierte Prozentzeichen, z.B.:

```
=SCORE  |piano| 
#
    1   |c    |
    1&  |%    |
    2   |%    |
    2&  |%    |
    3   |e    |
    3&  |%    |
    4   |%    |
    4&  |%    |
```

## Anschlagsstärke (Dynamik)

Die `Dynamik` (Anschlagsstärke, laut-leise) wird durch nachgestellte `+` und `-`-Zeichen angezeigt:

`c++++`  extrem laut       (fortississimo)
`c+++`   sehr laut         (fortissimo)
`c++`    laut              (forte)
`c+`     etwas laut        (mezzoforte)
`c `     normal
`c-`     etwas leise       (mezzopiano)
`c--`    leise             (piano)
`c---`   sehr leise        (pianissimo)
`c----`  extrem leise      (pianississimo)
`c=`     wie die Note davor


```
=SCORE |piano| 
#
    1  |c    |
    1& |d+   |
    2  |e    |
    2& |f-   |
#
    1  |g-   |
    1& |a-   |
    4  |d#'  |
    4& |f#   |
```

In diesem Beispiel haben die Noten `c`, `e`, `d#'` und `f#` die Anschlagsstärke *normal*, `d` ist *etwas laut* und `f,`g und `a` sind *leise*.

## Polyphonie

Wollen wir mehrere Töne gleichzeitig erklingen lassen, so benötigen wir mehrere `Spuren` - selbst, wenn das Instrument das gleiche ist, z.B.

```
=SCORE   |piano1|piano2| 
#
    1    |c     |      |
    1&   |      |c'    |
    2    |d     |      |
    2&   |      |c'    |
    3    |e     |      |
    3&   |      |c'    |
    4    |f     |      |
    4&   |      |c'    |
```

Hier spielt die Spur `piano1` die Töne `c`,`d`,`e` und `f` jeweils auf einem Schlag und als *Viertelnoten*, während `piano2` immer auf die Zeiten dazwischen (das "und") den Ton `c'` spielt, auch als *Viertelnoten*.
Auf dem Schlag 2 beginnt z.B. das `d` in `piano1`, während das `c'`, welches schon  auf `1&` in `piano2` begonnen hat, weiterklingt.

Wenn Töne genau parallel erklingen, also die *gleiche Startposition und die gleiche Länge* haben, kann man sie ausnahmsweise auch in eine Stimme schreiben. Dies ist bei *Parallelführungen* und *Akkordbegleitungen* interessant:

```
=SCORE  |piano-chords| 
#
      1 |(c e g)     |
      2 |(d f a)     |
#
      1 |*           |
```

Hier erklingen die Töne `c`,`e` und `g` gleichzeitig auf Schlag 1 und alle bis zum Schlag 2 (Viertelnoten), an dem sie von den Tönen `d`,`f`,`a` abgelöst werden, die bis zum Ende des Taktes erklingen (punktierte Halbenoten).

Mehrere Töne (oder auch allgemein Ereignisse) werden also als eine Gruppe parallel geführt, wenn in der Notation die Gruppe mit in Klammern steht und die Töne (bzw. Ereignisse) mit Leerzeichen getrennt sind.

Um Polyphonie auch innerhalb einer Spur zu ermöglichen, kann man mit einer alternativen Schreibweise auch explizit Start und Endpunkte
einer Note setzen. Um eine Startnote zu setzen, wird der Unterstrich _ ans Ende der Note gesetzt.
Um eine Note zu beenden, wird der Unterstrich vor die Note gesetzt. Jede auf diese Weise nicht beendete Note erklingt bis zum Schluss des Stückes:

```
=SCORE  |piano1|piano2| 
#
    1   |c_    |      |
    1&  |      |c'_   |
    2   |d     |      |
    2&  |      |d_'   |
    3   |e     |      |
    3&  |      |_c'   |
    4   |*     |      |
    4&  |_c    |_d'   |
```

Hier erklingen in Spur `piano2` ab Position `2& c'` und `d'` gleichzeitig, bis `c'` auf `3&` beendet wird und nur noch `d'` erklingt bis `4&`. In Spur `piano1` sind die *monophone* und die *polyphone* Schreibweise kombiniert: `c` erklingt von `1` bis `4&`, während `d` von `2` bis `3` klingt und `e` bis `4`.

Beide Schreibweisen können auch in "Akkordschreibweise" kombiniert werden:


```
=SCORE |piano-chords| 
#
    1  |(c_ e g_)   |
    2  |(: _c)      |
    3  |(f a)       |
    4  |(_g)        |
```

Hier wird `c` von `1` bis `2` gespielt. `e` würde normalerweise mit dem nächsten Akkord (also auf `2`) beendet werden.
Der Doppelpunkt verhindert dieses hier, so dass `e` bis `3` erklingt, wo es von `f` und `a` abeglöst wird, die bis
zum nächsten Akkord auf `4` klingen (obwohl dieser "Akkord" nur aus dem Notenende von `g` besteht). `g` erklingt durch
alles hindurch von `1` bis `4`.


## Gesangstexte

Bisher haben wir als `Ereignisse` nur Töne kennengelernt. Wollen wir jedoch Gesang notieren, 
so brauchen wir zusätzlich die Möglichkeit, Text den Tönen zuzuordnen. `Text` ist ein weiterer 
Ereignistyp und wird in doppelten Anführungsstrichen `"` geschrieben. Hier brauchen wir eine Gruppe, 
zusammen mit der dazugehörigen Note, z.B.

```
=SCORE |vocal        |
#
    1  |(c "Al-")    |
    1& |(d "-le")    |
    2  |(e "mei-")   |
    2& |(f "-ne")    |
    3  |(g "Ent-")   |
    4  |(g "-chen")  |
#
    1  |(a "schwim-")|
    1& |(a "-men")   |
    2  |(a "auf")    |
    2& |(a "dem")    |
    3  |(g "See")    |
```

Selbstverständlich kann der Text auch in einer eigenen Spur notiert werden, die den gleichen MIDI-Kanal 
zugewiesen bekommt und das ist wahrscheinlich auch besser lesbar:

```
TRACK   | vocal-notes | vocal-lyrics |
Channel | 1           | 1            |

=SCORE  |vocal-notes|vocal-lyrics|
#
    1   |c          |  "Al-"     |
    1&  |d          |  "-le"     |
    2   |e          |  "mei-"    |
    2&  |f          |  "-ne"     |
    3   |g          |  "Ent-"    |
    4   |g          |  "-chen"   |
#
    1   |a          |  "schwim-" |
    1&  |a          |  "-men"    |
    2   |a          |  "auf"     |
    2&  |a          |  "dem"     |
    3   |g          |  "See"     |
```

Man kann die Texte auch in eine separate Tabelle schreiben und mit den Noten verbinden.

## Kommentare

Während des Notierens und Komponierens kommt immer mal wieder die Situation auf, dass man Ideen 
ausprobiert und verwirft oder manchmal noch nicht genau weiß, ob man sie verwerfen will, oder sich 
später doch noch anders entscheidet.

In `MuSkel` hat man hierzu die Möglichkeit des "auskommentierens". Das bedeutet: jede Zeile, die mit zwei Querstrichen beginnt `//` ist ein Kommentar und damit reiner Text ohne besondere syntaktische Bedeutung.

Man kann dies Nutzen, um einzelne Stellen eines Stücks zu "deaktivieren" bzw. ignorieren, ohne sie durch Durchstreichen unleserlich zu machen:

```
=
      |piano| 
    1 |c    |
//     &|d+   |
    2 |e    |
//     &|f--  |

    1 |g    |
     &|a=   |
    4 |d#'  |
     &|f#   |
```

Im obigen Beispiel sind das `d` auf `1&`, sowie das `f` auf `2&` "auskommentiert" und werden damit deaktiviert.
Das betrifft auch alle Computerverarbeitung, wie etwa den `SMF-Export` und wird von der Syntaxhervorhebung entsprechend gekennzeichnet.

Dennoch kann man die auskommentierten Zeilen weiterhin lesen und verstehen:
Sie könnten als Gedankenstütze für eine Variation oder Alternative dienen, die man vielleicht zu einem späteren Zeitpunkt im Stück verwenden will.

Selbstverständlich kann man Kommentare aber auch verwenden, um Hinweise und Notizen zu machen, z.B.

```
=
      |piano| 
// besser im 3/4 Takt? überarbeiten!
    1 |c    |
     &|d+   |
    2 |e    |
     &|f--  |

    1 |g    |
     &|a=   |
    4 |d#'  |
     &|f#   |
```

Man kann auch mehrere Zeilen zugleich auskommentieren. Dies ist vor allem am Computer interessant, da dieser durch Syntaxhervorhebung klar zeigen kann, in welchem Bereich der Kommentar vorliegt. Auf analogem Wege können mehrzeilige Kommentare schnell unübersichtlich und verwirrend werden.

Der mehrzeilige Kommentar beginnt mit der Zeile, die mit Querstrich und Sternchen beginnt `/*` und endet mit der Zeile, die mit Sternchen und Querstrich beginnt `*/`

```
=
      |piano| 
    1 |c    |
     &|d+   |
/*
    2 |e    |
     &|f--  |

    1 |g    |
     &|a=   |
*/
    4 |d#'  |
     &|f#   |
```

Hier ist der Bereich von Schlag 2 des ersten Taktes bis Schlag 4 des nächsten Taktes auskommentiert.

Kommentare können selbstverständlich auch im Kopfbereich verwendet werden.

## Eigenschaften eines Stücks im Kopfbereich

Zuletzt schauen wir uns an, wie Eigenschaften des Stückes im Kopfbereich notiert werden. Die Syntax ist sehr einfach: Je Eigenschaften wir eine neue Zeile benötigt, die mit einem `@` anfängt, gefolgt von dem Namen der Eigenschaften, gefolgt von einem Doppelpunkt `:`. Nach dem Doppelpunkt
folgt der Wert der Eigenschaft. Hier ein Beispiel:

```
@title:   Demostück
@artist:  unknown

=
      |piano| 
    1 |c    |
     &|d+   |
    2 |e    |
     &|f--  |

    1 |g    |
     &|a=   |
    4 |d#'  |
     &|f#   |
```

Hier haben wir ein Stück mit den Eigenschaften `title` und `artist` im Kopfbereich und den Werten "Demostück" für `title` und "unknown" für `artist`.

Die Namen der Eigenschaften sind frei wählbar, ebenso die Anzahl. Es gibt jedoch ein paar Eigenschaften, die eine besondere Bedeutung haben und von entsprechenden Programmen ausgelesen und verwendet werden können:

`title`     Titel des Stücks
`artist`    Aufführender Musiker/Künstler
`composer`  Komponist
`date`      Datum der Komposition
`version`   Version (Nummer oder Name)
`copyright` Informationen zum Copyright
`genre`     musikalisches Genre
`tags`      Tags, durch Komma getrennt (z.B. für Suchmaschinen)

Bisher haben wir nur die Grundlagen von `MuSkel` kennengelernt, aber diese genügen bereits, um komplexe und interessante Musik zu notieren, aber es gibt noch viel mehr Möglichkeiten. Mit dem Computer, einem Editor
(am besten mit Hervorhebung der Syntax) und dem Hilfsprogram `muskel` ist es möglich, sehr komfortabel und schnell zu notieren und die Ergebnisse mittels `SMF-Export` sofort hörbar zu machen, in einer `DAW` weiterzuverarbeiten oder mit einem Notationsprogramm in die westliche Standardnotation zu überführen.
