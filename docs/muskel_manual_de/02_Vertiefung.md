# Taktwiederholungen und Bereiche

In MuSkel gibt es vielfältige Möglichkeiten zur Wiederholung.

Taktwiederholungen betreffen nur die Spur, in der sie definiert werden, während
Bereiche die Ereignisse aller Spuren in bestimmten Takten benennen und entsprechend
wiederholen können.

## Taktwiederholungen

Die einfachste Form der Taktwiederholung ist die einmalige Wiederholung des vorherigen Taktes
=
    |piano|flute| 
1   |c    |     |
2   |d    | e   |

1   |./.  | f   |
2   |     | g   |

1   |     | a   |
2   |     |     |

In diesem Beispiel wird in der piano Spur der erste Takt im zweiten wiederholt. Im dritten schweigt
die piano Spur. Die flute Spur spielt unabhängig davon ihre Töne.

Die Zeichenkette ./. drückt also die einmalige Wiederholung des unmittelbar vorangegangenen Taktes aus.

=
    |piano|flute| 
1   |c    |     |
2   |d    | e   |

1   |...  | f   |
2   |     | g   |

1   |     | a   |
2   |     |     |

1   |*    |     |

Durch Verwendung der Zeichenkette .n. wobei n eine Zahl ist, werden die letzten n-Take wiederholt. Zum Beispiel:

=
    |piano|flute| 
1   |f    |     |
2   |f    | e   |

1   |c    |     |
2   |d    | e   |

1   |.2.  | f   |
2   |     | g   |

1   |     | a   |
2   |     |     |

1&  |     | g   |

1   |     |     |

1   |     |     |

1   |d    |     |

dies entspricht

=
    |piano|flute| 
1   |f    |     |
2   |f    | e   |

1   |c    |     |
2   |d    | e   |

1   |f    | f   |
2   |f    | g   |

1   |c    | a   |
2   |d    |     |

1   |f    |     |
1&  |     | g   |
2   |f    |     |

1   |c    |     |
2   |d    |     |

1   |f    |     |
2   |f    |     |

1   |d    |     |

Die Wiederholung der Takte 1 und 2 in den Takten 3-7 wird unterbrochen, sobald im Takt 8 ein neues Ereignis auftritt.

## Bereiche

Will man Teile eines Stückes in allen Spuren wiederholen, so kann man Bereiche definieren und dann später wiederholen.

Bereiche umfassen immer vollständige Takte für alle Spuren. In der ersten notierten Zeile des Taktes wird
in der letzten Spalte der Name des Bereichs definiert, der bis zum nächsten Bereichsstart oder einer Bereichswiederholung reicht. Eine Bereichswiederholung ist eine Zeile, die in eckigen Klammern den Namen des
Bereichs beinhaltet, der wiederholt werden soll.

Beispiel

=
    |piano|flute| 
1   |c    |     |Area1
2   |d    | e   |

1   |f    |     |

1   |d'   | f   |Area2
2   |a'   | g   |
[Area2]
[Area1]
[Area1]

resultiert in

=
    |piano|flute| 
1   |c    |     |
2   |d    | e   |

1   |f    |     |

1   |d'   | f   |
2   |a'   | g   |

1   |d'   | f   |
2   |a'   | g   |

1   |c    |     |
2   |d    | e   |

1   |f    |     |

1   |c    |     |
2   |d    | e   |

1   |f    |     |

Der Bereich Area1 umfasst Takt1 und Takt2, da in Takt 3 Bereich Area2 startet. Area2 endet mit dem Ende des Taktes, da der nächste Takt aus einer Wiederholung besteht. Im Takt 4 wird Area2 wiederholt, in den Takt 5-6 und 7-8 wird Area1 wiederholt. 

Taktwiederholungen und Bereiche können kombiniert werden.

Beispiel

=
    |piano|flute| 
1   |c    |     |Area1
2   |d    | e   |

1   |./.  | f   |
[Area1]
[Area1]

wird zu


=
    |piano|flute| 
1   |c    |     |
2   |d    | e   |

1   |c    | f   |
2   |d    |     |

1   |c    |     |
2   |d    | e   |

1   |c    | f   |
2   |d    |     |

1   |c    |     |
2   |d    | e   |

1   |c    | f   |
2   |d    |     |

# Tempo, Tempowechsel und Tempoverläufe

Wie wir bereits bei den Grundlagen gesehen haben, ist das Standardtempo wenn nicht anders notiert 120 Schläge pro Minute (BPM). Eine Tempoangabe erfolgt immer mit einem Taktwechsel und wird durch ein vorangestelltes @ 
gekennzeichnet, z.B.

=
    |piano| 
1   |c    |
2   |d    |
@160
1   |c    |
2   |d    |

1   |d    |
2   |d    |

Hier hat der erste Takt das Tempo 120 BPM (Standard) und ab dem zweiten Takt liegt das Tempo bei 160 BPM.
Wollte man schon im ersten Takt das Tempo auf 160 BPM setzen, sähe es folgendermaßen aus:


=
    |piano| 
@160
1   |c    |
2   |d    |

1   |c    |
2   |d    |

1   |d    |
2   |d    |

Da nun auch ein Wechsel der Taktart mit einem Taktwechsel einhergeht, notiert man, so man beides will Taktart und
Tempo in der gleichen Zeile, wobei zuerst die Taktart und dann der Taktwechsel erfolgt:


=
    |piano| 
4/4@160
1   |c    |
2   |d    |

1   |c    |
2   |d    |
3/4@90
1   |d    |
2   |d    |

Hier sind die ersten beiden Takte 4/4 Takte mit dem Tempo 160 BPM und der dritte Takt ist ein 3/4 Takt mit dem Tempo 90 BPM.

Es sind auch Tempoverläufe möglich. Hierbei wird das Tempo zwischen zwei BPM Werten interpoliert, wobei jede 16tel
Note eine Tempoänderung erfolgt (wenn man zu SMF-Dateien exportiert):

=
    |piano| 
4/4@160~
1   |c    |
2   |d    |

1   |c    |
2   |d    |
3/4@90
1   |d    |
2   |d    |
@100
1   |c    |
2   |d    |

In diesem Beispiel drückt die einfache Tilde ~ hinter der BPM-Zahl im ersten Takt aus, dass ein linearer Tempoverlauf folgt bis zur nächsten BPM Zahl (in diesem Fall zu Beginn des 3. Taktes). Da die neue BPM Zahl (90)
niedriger ist, als die Ausgangszahl (160), wird eine lineare Verlangsamung von 160 BPM zu 90 BPM beim SMF-Export 
erzeugt und zwar so, dass für jedes 16tel der Takte 1 und 2 ein Tempowechsel mit einem Zwischenwert erfolgt.
Da bei auf die BPM Zahl 90 keine Tilde folgt, bleibt für den Takt 3 das Tempo statisch bei 90 BPM und wechselt
abrupt im Takt 4 auf 100 BPM.

=
    |piano| 
4/4@160~
1   |c    |
2   |d    |

1   |c    |
2   |d    |
3/4@190~~
1   |d    |
2   |d    |
@100
1   |c    |
2   |d    |

Werden statt einer Tilde, zwei Tilden verwendet, so wird der Tempoverlauf entlang einer Exponentilalfunktion gebildet. Im Beispiel steigert sich das Tempo linear von 160 BPM auf 190 BPM, um dann exponentiell auf 100 BPM abzufallen.


# Glissandi

Glissando nennt man ein kontinuierlichen Tonhöhenverlauf zwischen zwei Noten.
MuSkel unterstützt linear und exponentielle Glissandi, analog zu den Tempoverläufen.

Die Startnote eines linearen Glissandos erhält eine Tilde angehängt. Das Glissando endet
dann bei der nächsten Note derselben Spur.

Beispiel:

=
      |violin|
    1 |c'~   |  
    2 |c"    |  

Hier startet das Glissando auf dem 1.Schlag beim eingestrichenen c und geht bis zum zweigestrichenen c hoch,
welches auf dem 2. Schlag erreicht wird.

Das Glissando kann auch starten, nach dem die Note bereits erklingt, dies wird durch die freistehende Tilde
ausgedrückt:

=
      |violin|
    1 |c'    |  
     &|~     |  
    2 |c"    |  

Hier wird auf dem ersten Schlag das c' gespielt und in der Tonhöhe für ein Achtel behalten, dann geht es in
ein Glissando über bis zum c" welches auf dem 2. Schlag erreicht wird.

Durch die doppelte Tilde wird das exponentielle Glissando notiert:

=
      |violin|
    1 |c'    |  
     &|~~    |  
    2 |c"~~  |  
    4 |a'    |
    
Hier startet das exponentielle Glissando auf 1&, erreicht c" auf 2 und sinkt dann exponentiell auf a', welches
auf 4 erreicht wird und bis zum Ende des Taktes andauert (ein Viertel).

Beim SMF-Export werden die Glissandi als Pitchbend-Messages übersetzt. Diese stimmen aber nur, wenn der Pitchbend-Bereich (Pitchbend-Range) mit demjenigen des MIDI-Klangerzeugers übereinstimmt. Weitere Informationen dazu finden
sich im Kapitel 3 unter Glissandi und Kontrollerverläufe.  

# Skalennotation

## Skalenauswahl

Ein Skalenwechsel erfolgt immer mit einem Taktwechsel, ähnlich wie z.B. ein Tempowechsel oder ein Wechsel der Taktart. Folgende Skalen sind fest einprogrammiert und können direkt ausgewählt werden (für das eigene Erstellen von Skalen, siehe Kapitel 5 Skalen und Stimmungssysteme):

major      Dur
minor      Moll

sowie die alt-griechischen Modi (sog. Kirchentonarten)
dorian     Dorisch
phrygian   Phrygisch
locrian    Lokrisch
lydian     Lydisch
mixolydian Mixolydisch
ionian     Ionisch (wie Dur)
aeolian    Äolisch (wie Moll)

Die Verwendung erfolgt durch Voranstellen des Backslashs vor dem Namen der Tonart, gefolgt von einem Accent Circonflex ("Dach") ^, gefolgt von der Note, die die erste Stufe bildet, z.B.:

=
    |piano| 
\minor^a'
1   |     |

1   |     |

Hier haben alle Takte die Skala a-Moll, wobei der Grundton das eingestrichene a ist.

=
    |piano| 
1   |     |
\minor^g'
1   |     |

In diesem Beispiel hat der erste Takt die voreingestellte Skala (c-Dur auf eingestrichenem c) und
der zweite Takt g-Moll auf eingestrichenem g.

Die Skalen gelten für alle Spuren.
Ein Taktwechsel kann neben einem Skalenwechsel ja bekanntlich auch das Tempo und die Taktart wechseln.
In Kombination wird der Skalenwechsel in die Taktzeile als letztes geschrieben, also


=
    |piano| 
1   |     |
6/8@190\minor^g'
1   |     |

bzw.

=
    |piano| 
1   |     |
6/8\minor^g'
1   |     |


## Skalennoten

Der Grund, eine Skala zu nutzen liegt darin, nur skaleneigene Töne verwenden zu wollen. Hier spricht man auch
von Stufen, wobei die erste Stufe der Grundton der Skala ist.

Innerhalb einer Spur können wir nun eine Stufe einer Skala verwenden.
Hierzu verwenden wir den Accent Circonflex (das "Dach") ^ gefolgt von der Stufe in der Skala.

^1 ist also der Grundton (=die erste Stufe) der Skala, ^2 die zweite usw. 

Ein Skala kann aber auch "abwärts" gegangen werden.
Dies wird durch ein Minuszeichen ausgedrückt:

^-1 ist also eine Stufe abwärts vom Grundton der Skala, ^-2 zwei Stufen abwärts etc.

^0 darf in diesem Zusammenhang nicht verwendet werden und hat eine spezielle Bedeutung in Verbindung mit
Templates.

Beispiel

=
      |vocal-notes|vocal-lyrics|
\major^c
    1 |^1         |  "Al-"     |
     &|^2         |  "-le"     |
    2 |^3         |  "mei-"    |
     &|^4         |  "-ne"     |
    3 |^5         |  "Ent-"    |
    4 |^5         |  "-chen"   |

    1 |^6         |  "schwim-" |
     &|^6         |  "-men"    |
    2 |^6         |  "auf"     |
     &|^6         |  "dem"     |
    3 |^5         |  "See"     |

Dies ist die Entsprechung zum "Alle meine Entchen" aus dem Grundlagenkapitel. 

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


Der Vorteil der Skalennotation ist, dass wir das Stück leicht in eine andere Tonart verschieben (transponieren) können, wir müssen nur den Grundton der Skala verändern - der Rest kann gleich bleiben. Auch ist es einfach, z.B. die Skala nach Moll zu verändern und dem Stück damit einen anderen Charakter zu geben.

Auch in der Skalennotation stehen die gewohnten Dynamik und Stakkato, sowie Glissandozeichen zur Verfügung.

=
      |vocal-notes|vocal-lyrics|
\major^c
    1 |^1++~~     |  "Al-"     |
     &|^2+:       |  "-le"     |
    2 |^3--~      |  "mei-"    |
     &|^4*:       |  "-ne"     |
    3 |^5         |  "Ent-"    |
    4 |^5         |  "-chen"   |
    
Eine Besonderheit bei der Skalennotation ist die Möglichkeit, hinter die Skalennote eine Raute # zu hängen und
damit den resultierenden Ton um einen Halbton zu erhöhen. Da die meisten (und alle "eingebauten") Skalen aus 
Halb- und Ganztonschritten bestehen, ist es somit möglich, auch skalenfremde Töne zu erreichen, ohne die Skalennotation zu verlassen (und die damit einhergehenden Vorteile, die einfache Transposition etc.):

=
      |vocal-notes|vocal-lyrics|
\major^c
    3 |^5#        |  "Ent-"    |
    4 |^5         |  "-chen"   |

resultiert hier in

=
      |vocal-notes|vocal-lyrics|
    3 |g#         |  "Ent-"    |
    4 |g          |  "-chen"   |

Die Raute wird, wie in der normalen Notation auch vor etwaigen Dynamik-,Glissando- oder Stakkato-Zeichen notiert,
also


=
      |vocal-notes|vocal-lyrics|
\major^c
    3 |^5#++~     |  "Ent-"    |
    4 |^5         |  "-chen"   |
 
# Triolen, Quintolen und andere N-Tolen

Abgesehen von den bereits in den Grundlagen besprochenen Zeitaufteilungen bis hin zu 32tel Noten, sind Mehrfachteilungen dieser Zeitwerte in etwa Triolen (durch 3) oder Quintolen (durch 5) üblich. Generell
sind alle denkbaren Teilungen (N-Tolen) auch in der westlichen Standardnotation erlaubt, wenn auch nicht immer leicht zu lesen.

Bei MuSkel erfolgt die Notation recht einfach. Eine N-Tole wird als Einheit aufgefasst, welche zu einem bestimmten
Zeitpunkt gesetzt wird, so wie andere Noten auch. Das bedeutet, eine N-Tole kann auch nur auf "geraden" Taktteilungszeiten, wie Schläge (Viertel), Achtel, Sechzehntel und 32tel Zeiten im Takt plaziert werden, was es einfach macht, N-Tolen als Ersatz für einen entsprechenden Zeitwert einer normalen Note aufzufassen.

Die Teilung der N-Tole kommt durch die Anzahl der Elemente zustande, d.h. sie wird (z.B. beim SMF-Export) automatisch berechnet und muss nicht angegeben werden. "Lücken" werden durch entsprechende Pausenzeichen gesetzt.

In einem Punkt unterscheiden sich jedoch N-Tolen von sonstigen Noten: Während sonstige Noten immer dann enden, wenn die nächste Note, oder eine Pause, oder ein Taktwechsel folgt und daher die Länge der Note nicht bei der
Note angegeben wird, muss die N-Tole ihr Ende "kennen", um die Startpunkt ihrer Elemente zu bestimmen.

Schauen wir uns ein Beispiel an:

=
      |piano   | 
    1 |c       |
    2 |{*,f,g}3|

Hier haben wir ein c auf Schlag 1, welches bis zum Schlag 2 dauert (1 Viertel), gefolgt von einer Triole die aus den drei Elementen Pause, f und g besteht (innerhalb der geschweiften Klammer) und bis zum 3. Schlag dauert (Positionsangabe nach den geschweiften Klammern). Damit ist die Gesamtlänge der Triole von Schlag 2 bis Schlag 3 = 1 Schlag (Viertel) und diese Gesamtlänge wird auf drei Elemente aufgeteilt. Damit dauert jedes Element 1/12 vom 4/4 Takt. Da das erste Element eine Pause ist, beginnt das f auf 1/4 + 1/12 = 4/12 = 1/3 nach dem Beginn des Taktes und das g auf 1/4 + 2/12 = 5/12 nach dem Beginn des Taktes.

Mathematisch mag es etwas kompliziert erscheinen, aber intuitiv ist es leicht erfassbar, wenn man sich bewusst macht, dass die Gesamtdauer eben ein Viertel beträgt und damit genauso lang ist, wie das c, nur in drei gleiche 
Zeiteinheiten aufgeteilt.

Diese Notation wir auch nicht komplizierter, wenn wir sie mit komplexeren Dauern und mehreren Elementen kombinieren:


=
      |piano         | 
    2&|{*,f,g,*,g}3&.|

Hier hat die N-Tole eine 5-Teilung und dauert von 2& bis 3&.
Wenn es auch schwierig ist, sich vorzustellen, zu welchen Zeiten konkret die einzelnen Töne hier erscheinen,
Anfang und Ende sind einfach abzählbar und in der westlichen Standardnotation wäre eine solche Zeitaufteilung 
nicht notierbar, geweige denn lesbar.

In Kombination mit dem SMF-Export und dem Anhören des Ergebnisses am Computer ergeben sich dennoch interessante
Felder für Experimente, wenn man das Urteil dem Ohr überlässt.

Folgende wichtigen Regeln müssen beim Umgang mit N-Tolen beachtet werden:

1. Das Ende einer N-Tole kann nicht über den nächsten Takt hinausgehen. Gegebenenfalls muss der Takt vergrößert werden. Dennoch wird die letzte Note der N-Tole in den folgenden Takt übergebunden, wenn dieser mit einem : startet.

2. Die Endposition der N-Tole darf nicht vor der Startposition liegen.

3. Es darf in der gleichen Spur kein weiteres Ereignis zwischen Start- und Endposition der N-Tole liegen.

4. Eine Endposition muss immer angegeben werden, muss valide sein und innerhalb des Taktes liegen.
 
# Zufallselemente

In MuSkel gibt es zwei Möglichkeiten, Zufälligkeit ins Spiel zu bringen:
1. Über eine Wahrscheinlichkeit, mit der ein Ereignis ausgelöst wird.
2. Über eine Gruppe von Ereignissen, aus den per Zufall eines ausgewählt wird.

## Wahrscheinlichkeit

Die Wahrscheinlichkeit, dass ein Ereignis kommt, wird notiert, indem ein Fragezeichen, gefolgt von einer
Prozentzahl in eckigen Klammern notiert wird und dem Ereignis, welches mit dieser Wahrscheinlichkeit ausgelöst
wird vorangestellt wird.

Beispiel:

=
    | piano  | flute  | 
1   | c      |        |
2   | ?[30]d | e      |
3   |        | ?[80]f |

Hier wird in piano auf Schlag mit 30%iger Wahrscheinlichkeit die Note d gespielt.
Interessanterweise wird das vorangegangene c nur dann auf Schlag 2 enden, wenn das d folgt, ansonsten
erklingt es weiter bis zum Ende des Taktes.

Entsprechend wird in flute mit 80%iger Wahrscheinlichkeit auf Schlag 3 das f auf das e folgen. Ansonsten
klingt das e bis zum Ende des Taktes weiter.

In Kombination mit Takt- und Bereichswiederholungen können so interessante Variationen entstehen, insbesondere,
wenn der SMF-Export genutzt wird.

## Zufällig Auswahl

Wenn auf das Fragezeichen eine Klammer folgt, so wird zufällig zwischen den Elementen in der Klammer mit
gleicher Wahrscheinlichkeit ausgewählt.


Beispiel:

=
    | piano    | flute        | 
1   | c        |              |
2   | ?(d,*,*) | e            |
3   |          | ?(f,f,f,f,*) |
    
In piano endet das c in jedem Fall auf dem 2. Schlag. Entweder durch eine nachfolgende Pause, oder durch
das nachfolgende d, wobei die Chance, dass eine Pause folgt doppelt so groß ist, wie die Chance, dass das
d folgt. Bei flute ist die Chance auf das f 4x so groß, wie die Chance auf die Pause.


# Vergleich von MuSkel zur westlichen Standardnotation

TODO