# größere Überarbeitung

- bei scale notes sind " und ' als oktavierungszeichen aktiv. d.h. wenn die scalenote > 0 wird aufwärts oktaviert und wenn sie < 0, wird abwärts oktaviert

- einzelne Zahlen sind die Position in der Scala (also ohne vorangehendes Dach ^)
- das Dach kann an eine Zahl drangehängt werden, gefolgt von einer weiteren Zahl, die die Verschiebung
  um die Anzahl an Stufen ausdrückt (wichtig: wir zählen die Stufen von der 1 ausgehend: 1: 1.Stufe, 1: 2. Stufe etc.)

    1^2

  wäre die Stufe 1, verschoben um zwei Stufen aufwärts, also Stufe 3

    3^-2

  wäre Stufe 3, verschoben um zwei Stufen abwärts, also Stufe 1

    -2^4

  wäre Stufe 3

    6^0

  wäre Stufe 6

- damit kann man auch in Patterns das "einhängen" auf bestimmten Stufen ermöglichen, z.B.

    =pt  |      | 1     | 2     |
    #
    1    | #1   | =.(1) | =.(2) |
    1&   | #1^2 |       |       |
    2    | #1^1 |       |       |
    2&   | #1^3 |       |       |
    3    | #1^2 |       |       |
    3&   | #1^4 |       |       |
    4    | #1^3 |       |       |
    4&   | #1^5 |       |       |



# nächste TODOs

- frei stehende Plus- und Minuszeichen erlauben "aufgepropfte" Dynamik/Dynamikverläufe, z.B.

    =pt  | piano |
    #
      1  | c'++  |
      2  | d-    |

    =SCORE | piano       | piano-dyn |
    #
      1    | =pt.piano%4 | +++       |
    #
      2    |             | --        |
    #                   // exponentieller Dynamikverlauf
      1    |             | =~~       |
    #
      1    |             | ++        |
    
Die Plus- und Minuszeichen werden einfach auf die darunterliegende Dynamik "aufaddiert".

- Tempoverläufe scheinen nicht zu funktionieren

- make a scale out of a multiitem:

    =SCORE    | p |
    # \(c e g)  // below comes the g
      1       | 2 |
    
    

- allow [] for taking a part of a multiitem, e.g. (C e g c')[0] takes C, (C e g c')[1;3] takes (e c'),  
  (C e g c')[1:3] takes (e g c'). Also do it for shortcuts/tokens
  
- $$save(key,what...) erlaubt speichern von Werten in einer Token-Tabelle. Key ist .table.token.col von 
  einer Tokentabelle, die existieren muss.
  bei .table.token. wird die spalte der aktuellen Spalte verwendet. what kann folgendes sein:
  - ^1 ^2 etc. Scalenwert der aktuellen Scala in dieser Spalte in dieser Zeitposition
  - CC(12) der aktuelle Controllerwert für den Controller 12 in dieser Spalte in dieser Zeitposition
  - PB der aktuelle Pitchbendwert in dieser Spalte in dieser Zeitposition
  - AT der aktuelle Aftertouchwert in dieser Spalte in dieser Zeitposition
  - PT(12) der aktuelle Polyaftertouchwert für Taste 12 in dieser Spalte in dieser Zeitposition
  - % die aktuelle Note (letzte geschlagene Note oder Pause)
  gesetzt werden die platzhalter. wenn mehr platzhalter vorhanden sind, als werte, werden die überschüssigen parameter 
  beim aufruf gesetzt, z.B. .table.token.col(c')
  auf die gleiche Weise kann auch in Pattern tabellen gespeichert werden. Hierbei ist der key dann =pattern.col

- wir brauchen eine komplette überarbeitung, um die abhängigkeiten aufzulösen: jedes item muss in einer closure funktion
  münden, die den aktuellen kontext (parameter, Zeitposition etc.) speichern und letztlich erst am Schluss aufgelöst wird
  (beim unrolling). das wird wahrscheinlich erstmal extrem komplex, aber es wird es ermöglichen, alles mit jedem zu kombinieren

- es fehlt die Möglichkeit, innerhalb der Score Tabelle die Tracks in den Spalten zu wechseln. Ein Spaltenwechsel impliziert immer auch das Ende der vorangehenden Spalte (d.h. eine Pause am Ende). Ein Spaltenwechsel besteht aus einer Zeile, die wieder mit =SCORE beginnt. Nur jene Spalten, die nicht leer sind, werden gewechselt.

- weitere Änderungen (gründlich überlegen): 
  - Das ^ vor den Stufen wegnehmen. Damit wären die nackten Zahlen Stufennoten
  - (ggf. Spezialsyntax für slicing wegnehmen und durch $slice funktion ersetzen)

- Aufruf des Templates mit geschweiften Klammern: Parameter werden nicht anhand ihrer Nummer ersetzt, sondern nach der Reihenfolge, in der sie im Template erscheinen (d.h. es müssen alle angegeben werden unabhängig von der Nummer). So kann man dann das gleiche Template mal mit vordefinierten Wiederholungen verwenden und mal diese Überschreiben. Ausserdem kann man auch Platzhalter ohne Nummer verwenden; diese werden dann fortlaufend nummeriert, man kann auch beides kombinieren.
Die Taktwiederholungszeichen kann man auch innerhalb der Parameter verwenden: 
 - ... wiederholt den letzten Parameter, bis alle Parameter ausgefüllt sind.
 - .n. wiederholt die letzten n Parameter, bis alle Parameter ausgefüllt sind

# Was fehlt mir bei muskel?

1. Die Möglichkeit, den aktuellen Ton zu hören und ändern zu können 
Wir brauchen die Möglichkeit, einen Sinuston zu spielen, basierend auf einer Zeilen und Spaltenposition.
Dann brauchen wir die Möglichkeit, aus einem Editor heraus einen Befehl aufzurufen und Dateiname, Zeile und Spalte anzugeben. Oder wir haben ein Sonderzeichen, dass beim Parsen zum abspielen führt (z.B. `==`)

2. Die Visualisierung der Melodie
Man könnte die Konvention haben, dass wenn ein Spaltenname auf `<` endet, die Spalte "expanded" ist, d.h. der Spaltenname beinhaltet nicht das `<`, aber die Anordnung der Töne spiegelt durch Padding den Tonhöhenverlauf wider (von links nach rechts mit minimalen Abständen).

3. Es gibt keine Möglichkeit, auf einfachem Wege Noten in Spalten "nach unten" oder "nach oben" zu schieben.
Auch hier könnte ein besonderes Zeichen helfen, wir müssten allerdings Anfang und Ende der zu verschiebenden Noten markieren können. z.B. `^+(` für _Start verschieben nach unten_ mit Menge an Pluszeichen = Anzahl der Verschiebepositionen und analog `^-(` für _Start verschieben nach oben_ und `^)` für das _Ende der letzten Verschiebegruppe_ (kann am Ende der Spalte weggelassen werden)

 
- wir brauchen bei den tracks einen ambitus: from: to: type: (ignore,fit,mirror)
  ignore entfernt noten, die nicht reinpassen, fit verwendet die nächste passende note der skala, mirror spiegelt die note
  octaven nach unten/oben (solange, bis es passt)

- dokumentation aktualisieren

- some improvements to make it faster to write:
  - check that position goes not beyond bar length

- transponieren mit Halbtönen erlauben, z.B. 
  ^1# transponiert um einen schritt und eine MIDInote aufwärts 
  ^1bb transponiert um einen schritt und zwei MIDInoten abwärts 
  ^2#3 transponiert um einen schritt und drei MIDInoten aufwärts
  ^#3 transponiert um drei MIDInoten aufwärts
  ^b3 transponiert um drei MIDInoten abwärts

- allow repeat last event sign to alter volume and scale mounting %++ etc
- add repeat last pattern %% also with modifiers, e.g. %%++< or %%^2-->
  or with command pipe, e.g. %%/$reverse(=)
- überschreiben nur mit velocity erlauben z.B. /++ erhöht die "darunterliegende Note" in der velocity um ++
- überschreiben nur mit transposition erlauben z.B. /^1 erhöht die "darunterliegende Note" um einen schritt

- check and fix randomness

    
- cleanup scalenotes: make nt.ScaleNote a *int8 and copy it when duplicating. when it is not set, it is no scale note. otherwise it is the step (counting from 0). parsing makes 1->0 2->1 etc but -1->-1 this should make scale calculations and transpositions easier.

- fix mounting patterns / scale notes within patterns
- allow tempo changes inside columns of main score, independant from main tempo with @130 etc. to set back to the main tempo, use
  just @. also relative tempo possible, e.g. @50%

later:
- convert csv to muskel
- export to csv
- export to xls for printing
- import from xls
- make converter a separate package, reading from stdin and writing to stdout
- make muskel binary read from stdin and write to stdout
- tests reanimieren

## weiteres / Komplexeres
- tuning
- timbres
- Ducking?

## erledigt

- allow token tables to have multiple columns and to address them like this
  .table.row.col
  this way we could have "linked" parameters like

.short | bass      | git   | voc |
on     | CC(7,127) | c++   | d++ |
off    | CC(7,0)   | *     | *   |

$$embed(.short.*)

=SCORE  | bass       | git   | voc   |
1       | .on.bass   | .on.  | .on.  |
2       | .off.bass  | .off. | .off. |
#
1       | ...       | ...    | ...    |
*3

in .on. the dot inside the square brackets refers to the name of the current column
in combination with track imports we could write it like this

TRACK  | bass | git | voc |
// bass and voc tracks will be created without having a col in score
import | git  |     | git |

.short | bass      | git   | voc |
on     | CC(7,127) | c++   | d++ |
off    | CC(7,0)   | *     | *   |

$$embed(.short.*)

=SCORE  | git   |
1       | .on.  |
2       | .off. |
#
1       | ...   |
*3


  - allow Ab Db ebb etc.
  3. allow first barsign (#) to be missing, if time signature is 4/4 and tempo 120  
  2. allow floats as position indicator, e.g. 1.5 (==1&) 2.25 (==2,)  3.75 (==3&.) 1,125 (==1;), but allow only the given
     numbers (or round to the next allowed number when parsing/formatting)
  1. change the dot to a comma, for postion: 2&, instead of 2&.
  3. allow a position without pipes and add the missing pipes
  4. when a column has the name - then it is a "skip column" and will be deleted when parsing/formatting (for every table)
- piping via $cmd()/$cmd()/...

- pattern aufrufsyntax bereinigen/erweitern:
  folgende Spaltennamen sind in patterns nicht zulässig: ".","~","-","*","+"
  
  =.a ist das pattern im gleichen sketch in der spalte a, alle patterns des gleichen sketches müssen über =~ angesprochen werden (und nicht über den Namen des sketches)
  =~.a ist das pattern in der spalte a im sketch der heißt, wie die aktuelle spalte
  =.~ ist nicht zulässig
  =aa.~ ist das pattern im sketch =aa in der spalte, die den gleichen namen trägt wie die aktuelle spalte
  =aa ist das pattern im sketch =aa, welches in der ersten spalte ist
  die letzte spalte, die als Namen mit "!" beginnt, wird als erste Spalte behandelt, gibt es keine solche spalte, wird die
  tatsächlich erste spalte genommen
- fix special tables: for each of those tables there can be multiple definitions which result in 
  expanding each other (TRACK, SCALE, TUNING, TIMBRE). overwrites are last wins.
- sketchspalte als loop einbetten (template muss auf ... oder .n. enden) und testen
- track eigenschaften vervollständigen und vernünftig formatieren (möglicherweise immer alle Eigenschaften expandieren beim Formatieren)
- endlich vernünftige namen finden für:
  - sketchtable:       Sketch    (Entwurf)
  - sketchcolumn:      Pattern   (Schablone)
  - shortcut:          Token     (Kürzel)
