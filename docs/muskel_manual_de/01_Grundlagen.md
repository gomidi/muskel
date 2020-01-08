# Aufbau eines MuSkel Sketches

Das Fundament der `MuSkel`-Sprache ist der `MuSkel Sketch`.
Mit Papier und Stift umfasst er eine oder mehrere fortlaufende (nummerierte) Seiten, in der digitalen Welt ist es eine oder mehrere Dateien, deren Namen auf `.mskl` endet und die im selben Verzeichnis gespeichert sind.

Ein `MuSkel Sketch` beschreibt ein einzelnes musikalisches Stück bzw. eine Komposition. Jeder `MuSkel Sketch` besteht aus einem `Kopfbereich` (header), in dem zeitunabhängige Eigenschaften des Stücks stehen und dem `Notationsbereich` (score), in dem zeitgebundene Ereignisse stehen.

`Kopfbereich` und `Notationsbereich` werden durch eine Zeile, die mit einem *Gleichheitszeichen* beginnt und nichts weiteres beinhaltet, abgetrennt.

Der einfachste mögliche `Kopfbereich` ist leer. Wir werden später die genauen Details des `Kopfbereichs` besprechen.

Der `Notationsbereich` ist eine einfache Tabelle. Die Zeitachse schreitet von oben nach unten fort und jede Zeile im `Notationsbereich` stellt einen bestimmten Zeitpunkt im Takt dar.

Die erste und die letzte Spalte haben eine besondere Funktion - alle dazwischenliegenden Spalten sind Spuren, in denen `Ereignisse` stattfinden können.

Ein `Ereignis` wäre z.B., das eine Note beginnt. **Alle Ereignisse
in einer Zeile finden gleichzeitig statt.**

Die erste Spalte ist speziell. Sie beinhaltet am Anfang des `Notationsbereiches` die Abkürzungen für Eigenschaften der Spuren, die in diesem Anfangsbereich festgelegt werden und danach die *Position im laufenden Takt* bzw.
*Eigenschaften des laufenden Taktes*.

Schauen wir uns den einfachsten möglichen `MuSkel Sketch` an:

```
=
    |piano| 
1   |     |
```

Das Gleichheitszeichen trennt den `Kopfbereich` vom `Notationsbereich` ab.
Der `Kopfbereich` ist leer. Der `Notationsbereich` besteht aus zwei Spalten und zwei Zeilen. Die erste Spalte in der ersten Zeile ist leer. Das signalisiert, *dass in den folgenden Spalten die Namen der Spuren erscheinen*. Es wird empfohlen diese Namen immer in die erste Zeile im Notationsbereich zu notieren. **Leere Spurennamen sind nicht erlaubt**.

In der Regel entspricht eine `Spur` einem Instrument bzw. einer `Stimme` eines Instrumentes. Hier wäre das Instrument wahrscheinlich ein Klavier, aber letztlich ist die Benennung der Spur frei wählbar.

Die zweite Zeile bezeichnet die Zeit-Position *1. Schlag im aktuellen Takt*. Die `1` in der ersten Spalte der zweiten Zeile bedeutet *1.Schlag* (Beat/Viertelnote) im Takt (wie gezählt wird).

Doch woher kommt der Takt? Wenn nichts anderes definiert wird, ist der Takt ein `4/4`-Takt und das Tempo liegt dann bei 120 Schlägen pro Minute (`BPM`).

Der obige `MuSkel Sketch` besteht aus einem Takt und einer Spur, auf der im Takt nichts passiert.

# Noten und Takte

So ohne `Ereignisse` ist es noch nicht wirklich spannend, also lassen wir etwas passieren:

```
=
    |piano| 
1   |c    |
2   |d    |
3   |e    |
4   |f    |

1   |g    |
```

Hier haben wir nun die Noten `c`,`d`,`e` und `f` auf den Schlägen `1`,`2`,`3` und `4` des ersten Takts und die Note `g` auf dem 1. Schlag des zweiten Taktes.

Zunächst fällt auf, dass wir einen neuen Takt erzeugen, in dem eine leere Zeile einfügen. Dies ist streng genommen ein "Spezialfall", der jedoch häufig vorkommt: Die Eigenschaften des Taktes ändern sich nicht und es soll der nächste Takt kommen.

Wollen wir jedoch die Taktart ändern, so verwenden wir die "Bruchnotation" mit Querstrich `/`:

```
=
      |piano| 
    1 |c    |
    2 |d    |
    3 |e    |
    4 |f    |
3/4
    1 |g    |
```

Hier wechselt der Takt zum `3/4`-Takt, d.h. der erste Takt ist ein `4/4`-Takt (Standard) und der Zweite ein `3/4`-Takt. Um die Lesbarkeit zu erhöhen, wurden die Bezeichnung der Schlagpositionen eingerückt. Leerzeichen bzw. freie Flächen können in MuSkel frei verteilt werden.

Das bedeutet auch, dass die Spalten "schief" sein können:

```
=
      |piano| 
    1  |c    |
    2   |d    |
    3  |e    |
    4    |f    |
3/4
    1   |g    |
```

Dies ist zwar nicht besonders leserlich, aber valide und kann, wenn wenig Zeit vorhanden ist, hilfreich sein. Verwendet man `MuSkel` am Computer, so kann man mit dem Programm `muskel` die Datei beobachten lassen, so dass das Programm die Spalten automatisch sauber anordnet, wenn man Änderungen speichert. In Verbindung mit einem Editor, der automatisch geänderte Dateien neu laden kann (z.B. `Vim`, `Sublime Text`), kann man somit sehr komfortabel arbeiten, da man sich nicht um die Spaltenanordnung kümmern muss.

# Dauer der Töne und Monophonie einer Stimme

Zurück zu dem Inhalt des `Notationsbereiches`. Wir erkennen nun die Startpunkte der Töne, aber wie definieren wir die Dauer einer Note? Eine Besonderheit von `MuSkel` ist, dass alle Spuren zunächst *einstimmig* (monophon) sind (zur Ausnahme kommen wir später), d.h. es erklingt pro `Spur` immer nur ein Ton zur gleichen Zeit und dieser endet, wenn der nächste
Ton erklingt. Also so ähnlich, wie bei der menschlichen Stimme, oder bei Blasinstrumenten.

Wenn wir wollen, dass ein Ton endet, ohne dass ein neuer Ton erklingt, so drücken wir dies durch das `Stille`-Zeichen `*` aus:

```
=
      |piano| 
    1 |c    |
    2 |*    |
    3 |e    |
    4 |f    |
3/4
    1 |g    |
```

Hier endet die Note `c` auf dem zweiten Schlag (dauert also ein `Viertel`-Note lang).

Zeilen im Takt, in denen keine `Ereignisse` stehen, können weggelassen werden, also würde hier

```
=
      |piano| 
    1 |c    |
    3 |e    |
    4 |f    |
3/4
    1 |g    |
```

das `c` erst enden, wenn auf dem 3.Schlag das `e` folgt (das `c` dauert also eine `Halbe`-Note lang).

Wollen wir, dass schon der erste Takt ein `3/4`-Takt wird, so schreiben wir:

```
=
      |piano| 
3/4
    1 |c    |
    3 |e    |

    1 |g    |
```

# Noten und Oktavlage

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
- kleines cis bzw des: `c#`  
- eingestrichenes cis bzw. des: `c#'` 
usw.

Die `Standardstimmung` ist die `gleichschwebende Stimmung`. Daher gilt `cis`= `des` usw. Abweichende Stimmungen können aber definiert werden.

# Zeitpositionen im Takt

Für die Zeitpositionen im Takt gibt es auch noch feinere Aufteilungen, als die Schläge (Viertel) im Takt. Teilen wir einen Schlag (Viertel), so bekommen wir zwei Achtel. Diese werden mit "und" gezählt und mit `&` geschrieben:

```
=
      |piano| 
    1 |c    |
    1&|c'   |
    2 |d    |
    2&|c'   |
    3 |e    |
    3&|c'   |
    4 |f    |
    4&|c'   |
```

Hier haben wir ein `c'` auf jedem "und" bzw `&` (auch genannt Offbeat). So verkürzen sich alle Notenlängen im Beispiel auf `Achtel`-Noten. Die Schlagposition bzw. Zahl vor dem `&` kann man auch weglassen - dann 
bezieht es sich auf den *vorangegangenen notierten Schlag*, also könnten wir es auch so schreiben:

```
=
      |piano| 
    1 |c    |
     &|c'   |
    2 |d    |
     &|c'   |
    3 |e    |
     &|c'   |
    4 |f    |
     &|c'   |
```

Fällt die erste Note im Takt auf ein "und", so muss allerdings der Schlag zwingend mitnotiert werden, z.B.:

```
=
      |piano| 
    2&|c'   |
```

Es gibt noch feinere Aufteilungen, wie 16tel Noten:

```
=
       |piano| 
    1  |c    |
      .|d'   |
     & |c'   |
     &.|d'   |
    2  |d    |
      .|d'   |
     & |c'   |
     &.|d'   |
    3  |e    |
      .|d'   |
     & |c'   |
     &.|d'   |
    4  |f    |
      .|d'   |
     & |c'   |
     &.|d'   |
```


oder 32tel Noten:

```
=
        |piano| 
    1   |c    |
       ;|e'   |
      . |d'   |
      .;|e'   |
     &  |c'   |
      &;|e'   |
     &. |d'   |
     &.;|e'   |
    2   |d    |
       ;|e'   |
      . |d'   |
      .;|e'   |
     &  |c'   |
      &;|e'   |
     &. |d'   |
     &.;|e'   |
```
    
# Stakkato

Will man kurze Stakkato-Noten schreiben, kann es schnell unübersichtlich werden, z.B.:

```
=
        |piano| 
    1   |c    |
       ;|*    |
      . |d'   |
      .;|*    |
     &  |c'   |
      &;|*    |
     &. |d'   |
     &.;|*    |
    2   |d    |
       ;|*    |
      . |d'   |
      .;|*    |
     &  |c'   |
      &;|*    |
     &. |d'   |
     &.;|*    |
```

Dafür gibt es eine spezielle Notation, die die Noten auf 64telverkürzt und damit Stakkato-Feeling erzeugt:

```
=
        |piano | 
    1   |c     |
      . |d'::: |
     &  |c'    |
     &. |d'::: |
    2   |d     |
      . |d'::: |
     &  |c'    |
     &. |d'::: |
```

Im obigen Beispiel wäre jedes `d'` eine 64tel, da die Note mit drei Doppelpunkten `:::` endet.
Bei zwei Doppelpunkten `::` wäre es eine 32tel und bei einem Doppelpunkt `:` eine 16tel.

# Überbindungen

Eine Note dauert bis die nächste Note oder eine Pause (Stille) notiert ist, **auch über Taktwechsel hinweg**:

```
=
        |piano| 
    1   |c    |

    1&  |d    |
```

Hier endet das `c` erst auf 1& des folgenden Taktes.

Wollen wir, dass es schon auf der 1 des folgenden Taktes endet, so müssen wir dort eine Pause setzen:

```
=
        |piano| 
    1   |c    |

    1   |*    |
    1&  |d    |
```

Wollen wir die letzte Note wiederholen, so verwenden wir das isolierte Prozentzeichen, z.B.:

```
=
      |piano| 
    1 |c    |
     &|%    |
    2 |%    |
     &|%    |
    3 |e    |
     &|%    |
    4 |%    |
     &|%    |
```

# Anschlagsstärke (Dynamik)

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
=
      |piano| 
    1 |c    |
     &|d+   |
    2 |e    |
     &|f-   |

    1 |g-   |
     &|a=   |
    4 |d#'  |
     &|f#   |
```

In diesem Beispiel haben die Noten `c`, `e`, `d#'` und `f#` die Anschlagsstärke *normal*, `d` ist *etwas laut* und `f,`g und `a` sind *leise*.

# Polyphonie

Wollen wir mehrere Töne gleichzeitig erklingen lassen, so benötigen wir mehrere `Spuren` - selbst, wenn das Instrument das gleiche ist, z.B.

```
=
      |piano1|piano2| 
    1 |c     |      |
     &|      |c'    |
    2 |d     |      |
     &|      |c'    |
    3 |e     |      |
     &|      |c'    |
    4 |f     |      |
     &|      |c'    |
```

Hier spielt die Spur `piano1` die Töne `c`,`d`,`e` und `f` jeweils auf einem Schlag und als *Viertelnoten*, während `piano2` immer auf die Zeiten dazwischen (das "und") den Ton `c'` spielt, auch als *Viertelnoten*.
Auf dem Schlag 2 beginnt z.B. das `d` in `piano1`, während das `c'`, welches schon  auf `1&` in `piano2` begonnen hat, weiterklingt.

Wenn Töne genau parallel erklingen, also die *gleiche Startposition und die gleiche Länge* haben, kann man sie ausnahmsweise auch in eine Stimme schreiben. Dies ist bei *Parallelführungen* und *Akkordbegleitungen* interessant:

```
=
      |piano-chords| 
    1 |*c*e*g*     |
    2 |*d*f*a*     |
```

Hier erklingen die Töne `c`,`e` und `g` gleichzeitig auf Schlag 1 und alle bis zum Schlag 2 (Viertelnoten), an dem sie von den Tönen `d`,`f`,`a` abgelöst werden, die bis zum Ende des Taktes erklingen (punktierte Halbenoten).

Mehrere Töne (oder auch allgemein Ereignisse) werden also als eine Gruppe parallel geführt, wenn in der Notation die Gruppe mit einem Sternchen `*` *beginnt und endet*, sowie das Sternchen zwischen den Tönen (bzw. Ereignissen) notiert wird.

Um Polyphonie auch innerhalb einer Spur zu ermöglichen, kann man mit einer alternativen Schreibweise auch explizit Start und Endpunkte
einer Note setzen. Um eine Startnote zu setzen, wird der Unterstrich _ ans Ende der Note gesetzt.
Um eine Note zu beenden, wird der Unterstrich vor die Note gesetzt. Jede auf diese Weise nicht beendete Note erklingt bis zum Schluss des Stückes:

```
=
      |piano1|piano2| 
    1 |c_    |      |
     &|      |c'_   |
    2 |d     |      |
     &|      |d_'   |
    3 |e     |      |
     &|      |_c'   |
    4 |*     |      |
     &|_c    |_d'   |
```
 
Hier erklingen in Spur piano2 ab Position 2& c' und d' gleichzeitig, bis c' auf 3& beendet wird und nur noch d' erklingt bis 4&.
In Spur piano1 sind die monophone und die polyphone Schreibweise kombiniert: c erklingt von 1 bis 4&, während d von 2 bis 3 klingt
und e bis 4.

Beide Schreibweisen können auch in "Akkordschreibweise" kombiniert werden:


```
=
      |piano-chords| 
    1 |*c_*e*g_*   |
    2 |*:*_c*      |
    3 |*f*a*       |
    4 |*_g*        |
```

Hier wird c von 1 bis 2 gespielt. e würde normalerweise mit dem nächsten Akkord (also auf 2) beendet werden.
Der Doppelpunkt verhindert dieses hier, so dass e bis 3 erklingt, wo es von f und a abeglöst wird, die bis
zum nächsten Akkord auf 4 klingen (obwohl dieser "Akkord" nur aus dem Notenende von g besteht). g erklingt durch
alles hindurch von 1 bis 4.


# Gesangstexte

Bisher haben wir als `Ereignisse` nur Töne kennengelernt. Wollen wir jedoch Gesang notieren, so brauchen wir zusätzlich die Möglichkeit, Text den Tönen zuzuordnen. `Text` ist ein weiterer Ereignistyp und wird in doppelten
Anführungsstrichen `"` geschrieben. Hier brauchen wir eine Gruppe, zusammen mit der dazugehörigen Note, z.B.

```
=
      |vocal        |
    1 |*c*"Al-"*    |
     &|*d*"-le"*    |
    2 |*e*"mei-"*   |
     &|*f*"-ne"*    |
    3 |*g*"Ent-"*   |
    4 |*g*"-chen"*  |

    1 |*a*"schwim-"*|
     &|*a*"-men"*   |
    2 |*a*"auf"*    |
     &|*a*"dem"*    |
    3 |*g*"See"*    |
```

Selbstverständlich kann der Text auch in einer eigenen Spur notiert werden und das ist wahrscheinlich auch besser lesbar:

```
=
      |vocal-notes|vocal-lyrics|
    1 |c          |  "Al-"     |
     &|d          |  "-le"     |
    2 |e          |  "mei-"    |
     &|f          |  "-ne"     |
    3 |g          |  "Ent-"    |
    4 |g          |  "-chen"   |

    1 |a          |  "schwim-" |
     &|a          |  "-men"    |
    2 |a          |  "auf"     |
     &|a          |  "dem"     |
    3 |g          |  "See"     |
```

# Kommentare

Während des Notierens und Komponierens kommt immer mal wieder die Situation auf, dass man Ideen ausprobiert und verwirft oder manchmal noch nicht genau weiß, ob man sie verwerfen will, oder sich später doch noch anders
entscheidet.

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

# Eigenschaften eines Stücks im Kopfbereich

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
