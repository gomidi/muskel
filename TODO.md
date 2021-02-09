# TODOs

### nur Ausschnitt exportieren

im Flow modus werden zusätzlich zu den ! score und spalten noch start und end positionen berücksichtig (in nicht flow werden sie 
einfach ignoriert)

wenn in einer Taktzeile auf die Raute # ein > folgt, so wird erst ab dieser Zeile in die MIDI-Datei exportiert. dieser Takt wird also zum startpunkt der MIDI-Datei. 

wenn es mehrere Zeilen gibt, bei denen auf die Raute ein > folgt, so zählt die letzte dieser Zeilen als Startpunkt

folgt ein _ auf eine Raute, so endet die MIDI-Datei dort. alle noch klingenden Noten werden dort beendet.

wenn es mehrere Stellen mit _ gibt, so zählt die erste Stelle mit _

### nur bestimmte Spuren exportieren

- es gibt eine neue Solo spalte für Tracks. wenn diese eine Zahl > null beinhaltet, so repräsentiert die Zahl die entsprechende 
  Solo Gruppe. 

- beim Aufruf kann dann als Option eine Sologruppe gewählt werden, welche dann exklusiv exportiert wird.


### parts/sprungmarken

- es wäre wahrscheinlich besser, statt von parts von sprungmarken zu sprechen und wirklich zu springen, d.h. die nachfolgenden sprungmarken werden weiter
geschrieben; solange, bis was neues kommt (neue sprungmarke oder was anderes).

- da sprungmarken sich immer auf einen Takt beziehen, sollte es eine syntax geben, um innerhalb des Taktes an eine bestimmte Stelle zu springen, z.B.
  #MARKE:1&
  oder 
  #MARKE:-2&  (würde bedeuten 5 Achtel vor #MARKE)

- auf ähnliche weise könnte man die patternwiederholung automatisch/implizit machen und .1. bedeutet dann: springe nach einem takt wieder an die stelle
  vor einem takt etc. die % zeichen hinter einem pattern fallen dann weg. ich würde dann gerne das % für die Wiederholung der Takte verwenden
  also % bzw %1 statt .1., %2 statt .2. usw. anbestimmte Stellen im vorangehenden Takt würde man dann mit %2:2& springen.
  dann könnte man den Punkt statt des % Zeichens verwenden, z.B. .+++ statt %+++ für das Spielen der letzten Note, aber auch .5 für das Spielen der
  der exakte Note, die fünf Noten vor der jetzigen gespielt wurde (gleiche Noten werden in der Sammlung berücksichtigt) und .!5 für das Spielen der verschiedenen 
  Note, die fünf Noten vor der vorherigen gespielt wurde (gleiche Noten werden in der Sammlung nicht berücksichtigt)
  um das gut gewährleisten zu können, werden die letzten 60 ungleichen und die letzten 50 gleichen Noten eines Tracks immer mitgeschleift.
  bei den Takten werden immer die letzten 60 Takte eines Tracks mitgeschleift.

- dann kommt dem patternende eine besondere bedeutung zu: es kann mitten im takt sein und dann wird im ersten takt an der gleichen stelle bei der 
  wiederholung fortgesetzt. EIN PATTERNENDE KANN EIN PATTERN AUCH KLEINER ALS EINEN TAKT MACHEN, DANN WIRD INNERHALB DES TAKTES WIEDERHOLT:

  z.B. wird aus

  | =mot |   |
  | #    |   |
  | 1    | a |
  | 1&   | g |
  | 2    | g |
  | 3    | _ |


  | =mot |   |
  | #    |   |
  | 1    | a |
  | 1&   | g |
  | 2    | g |
  | 3    | a |
  | 3&   | g |
  | 4    | g |

oder aus 

  | =mot |    |
  | #    |    |
  | 1    | #1 |
  | 1,   | #2 |
  | 1&   | #3 |
  | 1&,  | _  |

wird

  | =mot |    |
  | #    |    |
  | 1    | #1 |
  | 1,   | #2 |
  | 1&   | #3 |
  | 1&,  | #1 |
  | 2    | #2 |
  | 2,   | #3 |
  | 2&   | #1 |
  | 2&,  | #2 |
  | 3    | #3 |
  | 3,   | #1 |
  | 3&   | #2 |
  | 3&,  | #3 |
  | 4    | #1 |
  | 4,   | #2 |
  | 4&   | #3 |
  | 4&,  | #1 |


mit diesen Änderungen wird die pattern einbettung mit eckigen Klammern und die part-adressierung mit eckigen klammern überflüssig.
stattdessen kann man auf folgende Weise an Stellen im pattern springen:

=patt

springt in den ersten Takt des patterns =patt in die gleiche Zeitposition wie diejenige, an der wir uns befinden (synchron)

=patt#2

springt in den zweiten Takt des patterns =patt in die gleiche Zeitposition wie diejenige, an der wir uns befinden (synchron)

=patt#2:2&

springt in den zweiten Takt des patterns =patt in die Zeitposition 2& und verschiebt dadurch ggf. die tonpositionen im Takt, wenn wir uns
nicht auch in Zeitposition 2& befinden

dadurch wird =! obsolet und das ! kann für andere Zwecke verwendet werden.




### sprungmarken können mit pattern modifizierern manipuliert werden

$merge([chorus],[verse])/$reverse(=)

### extra euclid funktion zum plazieren der schläge auf platzhaltern

d.h. die eigentliche Positionen der euclidifizierung sind nur indizes, die in die platzhalter
positionen des übergebenen patterns eingesetzt werden. (euclidpositioner o.ä. als name)

ggf ist es viel geiler, ein Funktion $placeinto zu haben, die vom aktuellen pattern nur die events in der 
reihenfolge nimmt und diese als parameter an das zweite pattern übergibt und damit "platziert".
dann könnte man einfach die rückgabe der $euclid als erstes pattern "reinpipen".

### syntaxerweiterung für geschweifte klammern

wir brauchen eine möglichkeit, der geschweiften klammer mitteilen zu können, wie lang ihre gesamtlänge ist,
unabhängig von dem was folgt. damit sind dann einfache ersetzungen besser vorhersehbar möglich.
vorschlag wäre, an den anfang nach der geöffneten geschweiften Klammer ein gleichheiszeichen zu setzen, gefolgt von 
der gesamtdauer und einem doppeltpunkt, also z.B.

{=&:4,6}

oder 

{=0.5:c#,.drums.sn++}

die alte syntax wird ungültig und der gleiche effekt, wie bei der alten Syntax wird erzielt, indem zwischen dem : und dem = nichts
steht:

{=:a,c,d}

damit sind ersetzungen besser vorhersehbar

### syntaxerweiterung skalennoten

es sollte erlaubt sein, die iteration (in der regel oktave) separat von der note anzugeben
dazu wird als trenner der slash verwendet, also

1/3

bedeutet: 1. ton der skala, drei oktaven höher

4/-2

bedeutet 4. ton der skala, zwei oktaven niedriger

auf diese Weise könnte man die negativen zahlen als deprecated definieren und fehler werfen, wenn die zahl größer ist,
als die elemente der skala (erleichtert das schreiben und erkennen gleicher noten ohne rechnen zu müssen)

für die "außenseiter noten" kann man dann den vorangestellten backslash benutzen

\2/3

bedeutet dann: 2. Außenseiternote der skala, drei oktaven höher

bei der skaltendefinition kann man dabei den zweiten backslash benutzen

\(c e g)\(b d')^C

(scala besteht aus c e und g) mit den außenseiter noten b und d' und dem Grundton C.

der Grundton der aktuellen skala kann dann einfach mit 

^

aufgerufen werden und mit

^/3

um drei oktaven nach oben transponiert werden.



### mengenfunktionen für patterns mit platzhaltern

schnittmenge: positionen, die in pattern1 und pattern2 vorkommen
exclusive: positionen, die in pattern1 und nicht in pattern2 vorkommen
outside: positionen, die in pattern1, aber weder in pattern2 noch in pattern3 vorkommen

diese mengen funktionen sollten auch für scalen gelten (schnittmenge von skalen, exclusive etc.)

entsprechend könnten $sub, $sub7 etc die subdominante, subdominante mit Septime etc als scala zurückgeben

### scalen sollen mit "outlaws" noten haben können, die nicht zur scala gehören,
aber eine besondere characteristic haben (z.B. Bluenote) und zu bestimmten zielskalen
modulieren können.

dann braucht es noch eine syntax, um das aufzurufen, z.B.

$outlaw(1)  // hole den 1. outlaw aus der aktuellen scala (könnte bei einem Dur-Akkord die Septime sein)
$modulate(1,2) // moduliere mittels des 1. outlaws der aktuellen scala zur 2. zielskala des 1. outlaws
               // setzt also diese zielskala als aktuelle skala

### auch in Taktpositionen gibt es wiederholungen durch .

z.B.

| =SCORE | git |
| #      |     |
| 1      | a   |
| 1&     | .   |
| .      | .   |

ist das gleiche wie 

| =SCORE | git |
| #      |     |
| 1      | a   |
| 1&     | a   |
| 2      | a   |
| 2&     | a   |
| 3      | a   |
| 3&     | a   |
| 4      | a   |
| 4&     | a   |

und

| =patt |   |
| #     |   |
| 1     | # |
| 1&    | . |
| .     | . |

| =bass |          | 1        |
| #     |          |          |
| 1     | =patt(d) | =patt(b) |
| 2     | /f       |          |
| 3     | /a       | /c       |

ist das gleiche wie

| =bass |   | 1 |
| #     |   |   |
| 1     | d | b |
| 1&    | d | b |
| 2     | f | b |
| 2&    | f | b |
| 3     | a | c |
| 3&    | a | c |
| 4     | a | c |
| 4&    | a | c |


### es gibt eine neue Schreibweise, um tonlängen zu definieren. 

normalerweise klingt eine note bis zur nächsten note, es sei denn, sie hat :, ::, oder ::: am Ende.
jetzt erweitert sich diese Schreibweise, so dass eine Dauer angegeben werden kann:

a:& bedeutet, dass das a für eine Achtel note lang klingt (es sei denn, es wird vorher durch eine andere Note oder Pause unterbrochen)

wenn nun ohne note in eine zelle

:&

geschrieben wird, bedeutet das, dass alle normal notierten noten, die in der spalte folgen, maximal eine Achtelnote lang sind
d.h.

a+++

wäre max eine Achtel lang

| =SCORE | melody |
| #      |        |
| 1      | :&     |
| 1&     | a:::   |
| 2      | a      |
| 3      | b      |
| 3,     | *      |

wäre gleichbedeutend mit

| =SCORE | melody |
| #      |        |
| 1&     | a:::   |
| 2      | a      |
| 2&     | *      |
| 3      | b      |
| 3,     | *      |

in der alten Notation

:=

hebt die Definition wieder auf, sodass alle Noten ohne Doppelpunkte wieder bis zur nächsten klingen

::

setzt die Standardlänge auf ::

| =SCORE | melody |
| #      |        |
| 1      | ::     |
| 1&     | a      |
| 2      | a:=    |
| 3      | b      |
| 3,     | *      |

wäre gleichbedeutend mit

| =SCORE | melody |
| #      |        |
| 1&     | a::    |
| 2      | a      |
| 3      | b      |
| 3,     | *      |


### standard lautheit

wenn in einer Zelle freistehend ein oder mehrere

+

oder 

-

Zeichen stehen, so setzen sie die Standardlautstärke von Noten, die ansonsten keine Dynamikvorschrift haben

ein

=

setzt die Standardlautstärke wieder auf den Standardwert

wenn eine Note nicht der so im Kontext gesetzten Standardlautstärke entsprechen will, sondern die 
ursprüngliche Standardlautstärke haben will, so setzt man ein = ans Ende:

a=

wenn die Lautstärke eine "darunterliegenden Note" (von einer Pattern oder Taktwiederholung oder einer Sprungwiederholung)
überschrieben werden soll, geschieht dies an der entsprechenden Position mit vorangehendem Slash
also

/++
/---
/=


### ereignisse in taktwechsel zeile

es ist möglich, in der zeile des taktwechsels in die entsprechenden Spalten ereignisse zu schreiben.
dies dürfen jedoch nicht notenereignisse sein.
die so in den Taktwechsel geschriebenen ereignisse geschehen am Taktanfang, jedoch vor den ereignissen, die in die Taktposition 1 geschrieben sind.
hier können z.B. Programwechsel, Controllereinstellungen, Sprungmarken, Patterns, Skalen, Pitchbendereignisse etc. stehen.
sollen hier in der gleichen Zelle mehrere Ereignisse stehen, so müssen sie in runde Klammern gefasst sein und durch Leerzeichen getrennt:

(PB0 [VERSE1] PC9)

sie werden von links nach rechts geschrieben

### grundsätzlich gilt: alles klingt und bleibt gleich, bis es verändert wird: noten erklingen bis zur nächsten, pattern werden wiederholt bis
zum nächsten ereignis. letzte Töne bleiben gleich bis zum nächsten, sprünge wiederholen, taktpositionen, standartlängen bis zur nächsten usw.

 
### CSV export von OpenOffice als mögliches Eingabe-Format akzeptieren (CSV wird nur gelesen)

### pattern einbettung

- zunächst sollten mit eckigen klammern die takte ausgewählt werden

[1:4]
Takt zwei bis vier

das sollte für patterns, aber auch für wiederholungen und parts gelten:

[part][1:4]
=pattern[1:4]
.6.[1:4]

-dann sollte das normale einbetten mit 

=pattern

so einbetten, dass die einbettungsposition im takt als startpunkt
für die nächsten noten genommen wird (also bezogen auf den ersten Takt des patterns)

Beispiel

```
=bass |   |
#
 1    | c |
#
 1&   | d |
 2    | e |
#
 3    | g |
 4    | b |


=SCORE | bass        |
#
 2     | =bass[1:]%2 |
*3
```

resultiert in

```
=SCORE | bass  |
#
 2     | e     |
#
 3     | g     |
 4     | b     |
#
 1&    | d     |
 2     | e     |
#
 3     | g     |
 4     | b     |
```

- das einbetten mit =! ist wie gehabt (aber bezogen auf das, was nach [1:4] überig bleibt)

### pattern funktionen als on/off modifier pipe

- die idee ist, dass wir die events in einem track durch eine reihe von funktionen schicken, die sie modifizieren
  und die unabhängig von einander aktiviert und deaktiviert/verändert werden können
  also z.B. die töne werden transponiert, rückwärts gespielt und gespiegelt.
  dann könnte das rückwärtsspielen deaktiviert werden, wobei aber transponieren und spiegelung erhalten bleibt
  wichtig dabei ist, dass die dauern und wiederholungsraten
  der funktionen sich von denjenigen der töne unterscheiden können
  ich habe die Vermutung, dass es der beste weg sein könnte, wenn jede dieser funktionen in einem pattern steht und wir eine syntax haben,
  mit der die reihenfolge jener pattern verknüpft wird. die frage, wie die länge und wiederholungsrate bestimmt wird, ist noch offen
  die einfachste möglichkeit wäre, auch dass über pattern funktionen zu machen, z.B.


=transpose|                               |
#
  1       | =#/$bars(1,5)/$transpose(3)   |
*2
#
  1       |                               |
#
  1       | =#/$bars(5,6)/$transpose(5)   |
#
  1       | _                             |

 
=reverse  |                         |
#
  1       | =#/$slice(5,9)/$reverse |
#
  1       | ...                     |
#
  1       | =#/$slice(5,9)          |
#
  1       | _                       |


=mirror   |            |
#
  1       | =#/$mirror |
#
  1       | _          |


=melody   |         |
#
  1       | c       |
  1&      | e       |
  3       | f       |
  4&      | e       |
#
  1       | ...     |
#
  1       | _       |


=SCORE    |                                               |   
#
  1       | =melody%12/=transpose%2/=reverse%4/=mirror%12 |

zu ändern wären dann nur, dass bei der pipe auch pattern wie funktionen angegeben werden können, wobei dann =# durch den unique-namen 
eines temporär erstellten patterns, welches die aktuellen elemente beinhaltet, ersetzt wird und die parameter normal übergeben werden.

innerhalb der "modifikationspattern" können dann die jeweiligen funktionen aufgerufen, ersetzt werden etc.
wichtig ist, dass wir eine zusätzliche funktion $bars(from,to) haben, welche slicing basierend auf takten ermöglicht und 
$position(from,to), welche slicing basierend auf positionen im takt erlaubt (unabhängig von der taktart), z.B. $position(4&,7) kann auch auf einen
4/4 takt bezogen werden, würde dann auf der 3 des Folgetaktes enden.

zu überlegen wäre, ob =# nicht in jedem takt einfach die elemente des gleichen taktes übergibt, wodurch man sich das taktzählen und bars raussuchen sparen
könnte. allerdings ist es weniger klar und explizit und zudem unklar, was passiert, wenn die funktionen nicht auf der position 1 aufgerufen werden.
also erstmal besser nicht. (sonst vielleicht durch ein anderes symbol, z.B. =!#)

die länge ist ja durch die länge des patterns vorgegeben.

  
### externe takt eigenschaften importieren
- es wäre schön, wenn man globale takt eigenschaften, wie
  - partname
  - taktart
  - tempo
  - tonart
  - tempoverläufe
 in eine eigene datei auslagern könnte. dass wäre dann eine =SCORE tabelle mit irgendwelchen normalen spalten
wo dann die ganzen takteigenschaften normal notiert drin stehen
für den import würde man dann einfach in dem entsprechenden Takt (in der Regel im ersten) folgendes schreiben

=SCORE | voc | bass |
#'externalbars
  1    | c   | g    | 
#
  2    |     |      |

etc

damit würden die taktinformationen aus der datei externalbars.mskl und deren tabelle =SCORE übernommen werden bis
zu einem weiteren takt der eine externe datei referenziert.
sämtliche folgenden taktinformationen im zielscore überschreiben gleichartige der eingebundenen datei, also z.b.

inhalt von externalbars.mskl
=SCORE | something |
# 4/4 @120
#INTRO @60

inhalt von main.mskl 


=SCORE | voc | bass |
#'externalbars
  1    | c   | g    | 
#OUTRO
  2    |     |      |

hier wird für den zweiten Takt der partname OUTRO verwendet, das Tempo aber dennoch auf 60 bpm geändert

### andere bugs:
- wenn auf eine gruppenspalte eine normale spalte folgt, kann diese nicht ordentlich geparst werden (panic)
- zumindest unter windows können keine dateien von höheren verzeichnissen inkludiert werden


- außerdem wäre es nützlich, eine option zu haben, um in einer spalte die töne durch midi notes zu ersetzen, etwa, wenn der spaltenname auf # endet
- außerdem wäre es nützlich, eine option zu haben, um in einer spalte alle töne als verkürzt umzunotieren, etwa, wenn der spaltenname auf :, :: oder ::: endet
- außerdem wäre es nützlich, eine option zu haben, um in einer spalte alle töne als scalentöne umzunotieren, etwa, wenn der spaltenname auf ^ endet, zusätzlich muss eine scala per parameter der kommandozeile angegeben werden (dem fmt subcommand, etwa --scale=major^c'#)
- außerdem wäre es nützlich, eine option zu haben, um in einer spalte alle dynamikzeichen zu entfernen, etwa, wenn der spaltenname auf +- endet

### big table rewrite -> markdown tables

- Das einzig blöde ist, dass es keine elegante Schreibweise für colspans gibt. also kommentare vergrößern damit die erste spalte (wenn es nicht auskommentieren einer ganzen Zeile ist, da muss das kommentarzeichen nur nach der ersten pipe geschrieben werden) 
wichtig ist auch, dass ``\`\`\`` und `\~\~\~` am Anfang und Ende einer Zeile als Kommentar gewertet werden (damit auch Muskel-Code "zitiert" werden kann)
- Das freistehende @ markiert die Position, ab der im Play-Modus gespielt wird (bis zum nächsten freistehenden @ oder zum Ende). Das freistehende @ wird nur beachtet, wenn es im SCORE steht (oder in der Tabelle, die dem Program als Haupttabelle genannt wurde). Ist das freistehende @ in der ersten Spalte, wird der gesamte SCORE ab dem freistehenden @ bis zum nächsten @ gespielt (alle Spalten/Stimmen).
Steht das freistehende @ in einer anderen Spalte, wird nur diese Spalte gespielt (stummschaltung bei den Tracks bzw fehlende Midi-Kanäle werden natürlich respektiert).
- Außerdem brauchen wir noch einen Webserver, über den direkte MIDI-Verbindungen zu einer DAW ausgewählt werden können, so dass man auch andere Instrumente gut direkt bespielen kann.
- Da Programme, wie Typora keinen Formatierer brauchen, sollte es noch einen check-Modus geben, in dem die Datei nur geprüft wird und etwaige Fehler gemeldet werden (ohne, dass in die Datei geschrieben wird). Das kann dann auch wieder mit watch und SMF export und play kombiniert werden. (vielleicht passiert check auch automatisch, wenn fmt nicht extra gewählt wird).
- Schön wäre auch ein Tool, mit dem man direkt in MIDI aufnehmen kann und gleichzeitig in eine muskel-Datei konvertieren kann.
- Das interface könnte über einen Webserver laufen, wo man den MIDI-Input-Kanal und den Ausgang auf dem es erklingen soll auswählen kann und das Tempo und die Anzahl an Vorlauftakten des Klicks. - Außerdem kann man darüber dann die Aufnahme starten und beenden. Es wird immer an die gewählte Datei angehangen mit Vermerk des datums und der uhrzeit in einem kommentar. Möglicherweise erlaubt man auch eine Loop-Aufnahme über n Takte, so dass für jeden weiteren Durchlauf eine neue Spalte in der Tabelle angelegt wird. Auf jeden Fall wird in der Spalte der Tabelle der MIDI-Kanal erfasst, über den die Noten/CCs reinkamen, so dass man auch auf mehreren Kanälen gleichzeitig einspielen kann.
- Möglicherweise benutzt man dafür im Hintergrund midish.
- Um möglichst kompatibel zu bleiben, wird innerhalb einer Tabelle eine fehlendes vorangehendes Pipe-Zeichen ergänzt (solange die Tabelle nicht durch eine Leerzeile unterbrochen wurde). Lediglich die erste Zeile der Tabelle muss mit einem Pipesymbol beginnen.

### was noch fehlt bei den gruppenspalten und der include vereinheitlichung

- Parts von externen Score includes sind noch nicht implementiert
- Parts von externen ScoreCell includes sind noch nicht implementiert
- gruppenspalten die auf andere gruppenspalten referenzieren, funktionieren nicht (ist das nötig?)
- verheitlichung und überall nutzung der reference library, insbesondere der kontexte (auch zur validitätsprüfung), zur zeit unheitlicher mischmasch
- slicing ([]) von externen tokens funktioniert noch nicht und auch sämtliche anderen modifier ++ ^2 etc.
- parameter übergabe an externe tokens funktioniert noch nicht
- modifier von externen patterns funktionieren noch nicht
- parameterübergabe an externe und interne Score level includes funktioniert noch nicht
- parameterübergabe an externe File level includes funktioniert noch nicht
- dieses ganze parameter und modifiergedöns müsste auch vernünftig in die references bibliothek integriert werden.


### weiteres
- gleichheitszeichen frei machen, reset von dynamik mit +-



#### grundsätzliche überarbeitung:
  Obergruppe 
    EventSequence: zeitliche Abfolge von Items einer Spur
        - haben eine Callbackfunktion, die die events holt; sie können von patterns/tabellenspalten, 
          wiederholungen mit .n., Wiederholungen mit % oder partwiederholungen stammen oder von includes
        - auf sie können folgende funktionen angewendet werden:
          - sequencecommands
          - slicing mit []
          - dynamik change mit + und -
          - überschreiben mit /
          - diatonische transposition mit ^
          - chromatische transposition mit # und b
          - octavierungen mit " und '
          - startsynchronisierung mit !
          - zufallswahrscheinlichkeit mit xx%
          - zufällige auswahlmit %()
          - microtimingverschiebung mit > und <
        - sie haben eine adresse von der die events geholt werden, z.b.
          =a.b
          =.b
          =.
          %
          %2
          ...
          .2.
          [part]
          [=table.col.part]
          [=.col.part]
          [=.part]
        - sie haben einen kontext, über den Adressierung und Skalenauflösung funktioniert (Position, Tabelle, Spalte, Track)
        - sie können mit () und {} parametrisiert werden, z.B.
          %(c',d')
          .3.{:,d,e}
    ItemGroup: items, die zu einer Gruppe gehören, ohne zeitliche Anordnung
        - haben eine Callbackfunktion, die die items holt; sie können von shortcuts, Multiitems oder Skalen kommen
        - sie können durch nachstellen von ... in parameter aufgelöst werden
        - sie haben eine inhärente wiederholung
        - auf sie können folgende funktionen angewendet werden:
          - sequencecommands
          - slicing mit []
          - dynamik change mit + und -
          - diatonische transposition mit ^
          - chromatische transposition mit # und b
          - octavierungen mit " und '
          - zufallswahrscheinlichkeit mit xx%
          - zufällige auswahlmit %()
          - microtimingverschiebung mit > und <
        - sie haben eine adresse, von der die items geholt werden, z.B.
          .a.b
          .b
          %
          %2
          \minor^c
          \(c e g)
          (c e g)
        - sie haben einen kontext, über den Adressierung und Skalenauflösung funktioniert (Position, Tabelle, Spalte, Track)
        - sie können mit () und {} parametrisiert werden, z.B.
          .a.b(c',d')
          \minor^c{c,d,e}
    Parameter: werte, die einem command, einer itemgroup oder einer EventSequence zur parametrisierung übergeben werden
        - spezielle Sonderzeichen, wie : % und .n. die nur im kontext des Aufrufs gelten
        - jeder parameter ist entweder ein spezielles Sonderzeichen, oder ein einfacher String, oder ein Item, oder eine ItemGroup oder 
          eine EventSequence. in einem command kann getestet werden, ob ein parameter ein einfacher String, oder ein Item, oder eine ItemGroup oder 
          eine EventSequence ist (spezielle Sonderzeichen werden während des aufrufs ersetzt)        
        - können einfach in eine Itemgroup verwandelt werden, z.B.
        =x  | 1  | 2               |
        #
         1  | #1 | =.1((#1 #2 #3)) |
    commands: 
        - empfangen parameter und geben eine itemgroup oder eine EventSequence oder nichts zurück.
        - haben eine andere syntax: wenn sie eine EventSequence zurückgeben, beginnen sie mit $= wenn sie eine itemgroup zurückgeben, beginnen 
          sie mit $_, wenn sie eine tabelle zurückgeben mit $$ (z.B. $$include) wenn sie nichts zurückgeben mit $ (z.B. $embed).

- wir brauchen eine komplette überarbeitung, um die abhängigkeiten aufzulösen: jedes item muss in einer closure funktion
  münden, die den aktuellen kontext (parameter, Zeitposition etc.) speichern und letztlich erst am Schluss aufgelöst wird
  (beim unrolling). das wird wahrscheinlich erstmal extrem komplex, aber es wird es ermöglichen, alles mit jedem zu kombinieren

- frei stehende Plus- und Minuszeichen erlauben "aufgepropfte" Dynamik/Dynamikverläufe, z.B.

```muskel
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
      1    |             | +-~~      |
    #
      1    |             | ++        |
```

Die Plus- und Minuszeichen werden einfach auf die darunterliegende Dynamik "aufaddiert".


- allow [] for taking a part of a multiitem, e.g. (C e g c')[0] takes C, (C e g c')[1;3] takes (e c'),  
  (C e g c')[1:3] takes (e g c'). Also do it for shortcuts/tokens

- Aufruf des Templates mit geschweiften Klammern: Parameter werden nicht anhand ihrer Nummer ersetzt, sondern nach der Reihenfolge, in der sie im Template erscheinen (d.h. es müssen alle angegeben werden unabhängig von der Nummer). So kann man dann das gleiche Template mal mit vordefinierten Wiederholungen verwenden und mal diese Überschreiben. Ausserdem kann man auch Platzhalter ohne Nummer verwenden; diese werden dann fortlaufend nummeriert, man kann auch beides kombinieren.

- allow repeat last event sign to alter volume and scale mounting %++ etc
- add repeat last pattern %% also with modifiers, e.g. %%++< or %%^2-->
  or with command pipe, e.g. %%/$reverse(+-)
- überschreiben nur mit velocity erlauben z.B. /++ erhöht die "darunterliegende Note" in der velocity um ++
- überschreiben nur mit transposition erlauben z.B. /^1 erhöht die "darunterliegende Note" um einen schritt

- Die Taktwiederholungszeichen kann man auch innerhalb der Parameter verwenden: 
 - ... wiederholt den letzten Parameter, bis alle Parameter ausgefüllt sind.
 - .n. wiederholt die letzten n Parameter, bis alle Parameter ausgefüllt sind

- funktionen, die sich auf die spalte auswirken, z.B. 
   - relatives tempo zum haupttempo
   - geändertes midiprogram
   - geändertes delay

  syntax: $.() z.B. 
    $.delay(3/16)

## Was fehlt mir bei muskel?

1. Die Möglichkeit, den aktuellen Ton zu hören und ändern zu können 
Wir brauchen die Möglichkeit, einen Sinuston zu spielen, basierend auf einer Zeilen und Spaltenposition.
Dann brauchen wir die Möglichkeit, aus einem Editor heraus einen Befehl aufzurufen und Dateiname, Zeile und Spalte anzugeben. Oder wir haben ein Sonderzeichen, dass beim Parsen zum abspielen führt (z.B. `==`)

2. Die Visualisierung der Melodie
Man könnte die Konvention haben, dass wenn ein Spaltenname auf `<` endet, die Spalte "expanded" ist, d.h. der Spaltenname beinhaltet nicht das `<`, aber die Anordnung der Töne spiegelt durch Padding den Tonhöhenverlauf wider (von links nach rechts mit minimalen Abständen).
Oder (einfacher) mit smfimage.

3. Es gibt keine Möglichkeit, auf einfachem Wege Noten in Spalten "nach unten" oder "nach oben" zu schieben.
Auch hier könnte ein besonderes Zeichen helfen, wir müssten allerdings Anfang und Ende der zu verschiebenden Noten markieren können. z.B. `^+(` für _Start verschieben nach unten_ mit Menge an Pluszeichen = Anzahl der Verschiebepositionen und analog `^-(` für _Start verschieben nach oben_ und `^)` für das _Ende der letzten Verschiebegruppe_ (kann am Ende der Spalte weggelassen werden)

Letztlich können alle drei Features aber besser in einem eigens entwickelten Editor implementiert werden.

## weiteres

- dokumentation aktualisieren
- some improvements to make it faster to write:
  - check that position goes not beyond bar length
- check and fix randomness
- allow tempo changes inside columns of main score, independant from main tempo with @130 etc. 
  to set back to the main tempo, use just @. also relative tempo possible, e.g. @50%

## later
- convert csv to muskel
- export to csv
- export to xls for printing
- import from xls
- make converter a separate package, reading from stdin and writing to stdout
- make muskel binary read from stdin and write to stdout
- tests reanimieren
- es fehlt die Möglichkeit, innerhalb der Score Tabelle die Tracks in den Spalten zu wechseln. Ein Spaltenwechsel impliziert immer auch das Ende der vorangehenden Spalte (d.h. eine Pause am Ende). Ein Spaltenwechsel besteht aus einer Zeile, die wieder mit =SCORE beginnt. Nur jene Spalten, die nicht leer sind, werden gewechselt.
- (ggf. Spezialsyntax für slicing wegnehmen und durch $slice funktion ersetzen)
- $$save(key,what...) erlaubt speichern von Werten in einer Token-Tabelle. Key ist .table.token.col von 
  einer Tokentabelle, die existieren muss.
  bei .table.token. wird die spalte der aktuellen Spalte verwendet. what kann folgendes sein:
  - 1,2 etc. Scalenwert der aktuellen Scala in dieser Spalte in dieser Zeitposition
  - CC(12) der aktuelle Controllerwert für den Controller 12 in dieser Spalte in dieser Zeitposition
  - PB der aktuelle Pitchbendwert in dieser Spalte in dieser Zeitposition
  - AT der aktuelle Aftertouchwert in dieser Spalte in dieser Zeitposition
  - PT(12) der aktuelle Polyaftertouchwert für Taste 12 in dieser Spalte in dieser Zeitposition
  - % die aktuelle Note (letzte geschlagene Note oder Pause)
  gesetzt werden die platzhalter. wenn mehr platzhalter vorhanden sind, als werte, werden die überschüssigen parameter 
  beim aufruf gesetzt, z.B. .table.token.col(c')
  auf die gleiche Weise kann auch in Pattern tabellen gespeichert werden. Hierbei ist der key dann =pattern.col
- diatonische verläufe mittels =, z.B.

 ```muskel
    =SCORE | piano |
    #
      1    | 2=&   |
    #
      1    | 10    |
    
  ist das gleiche, wie

    =SCORE | piano |
    #
      1    | 2     |
      1&   | 3     |
      2    | 4     |
      2&   | 5     |
      3    | 6     |
      3&   | 7     |
      4    | 8     |
      4&   | 9     |
    #
      1    | 10    |
 ```

  voraussetzung dazu ist, dass das Lautstärkeresetzeichen von = auf +- geändert wird. (nach einer note gleicht sich das sowieso aus,
  aber bei applizierung auf patterns etc. ist es wichtig)

- mehrfache chromatische transposition erlauben, z.B. 
  1^1# transponiert um einen schritt und eine MIDInote aufwärts 
  1^1bb transponiert um einen schritt und zwei MIDInoten abwärts 
  1^2#3 transponiert um einen schritt und drei MIDInoten aufwärts
  1#3 transponiert um drei MIDInoten aufwärts
  1b3 transponiert um drei MIDInoten abwärts

## Komplexeres
- tuning
- timbres
- Ducking?


# erledigt

- we need a formatting option to delete empty lines

### big table rewrite -> markdown tables
- erste spalte von tabelle beginnt immer mit `|` die zweite zeile kann `|---|---|` etc enthalten. 
- beim formatieren wird sie immer so umgeschrieben, dass sie `|---|---|` enthält. 
- Damit werden muskel Dateien automatisch zu markdown Dateien und die muskel-tabellen zu markdown tabellen. 
- Damit können sie mit jedem markdown-Editor bearbeitet werden (und in schönes HTML oder PDF umgewandelt werden)
- ein schöner Editor dafür ist Typora.
- Alles, was nicht in Tabellen steht, ist Kommentar. (außer den `$includes`, `$embed` etc.) 
- Damit kann man eine Komposition als normale markdown-Datei schreiben mit Überschriften etc. und ganz normal formatieren und trotzdem eine Komposition drinnen enthalten haben. Allerdings darf in den Kompositionstabellen nicht "formatiert" werden.

### für die midi bibliothek
vielleicht mehr für smftrack: 
funktion, die in einem track überlappende noten beseitigt / den track monophon macht (innerhalb des gleichen channels) und 
eine weitere funktion, die prüft, ob ein track monophon ist (innerhalb des gleichen channels).
macht natürlich nur sinn, wenn in einem track keine keyswitches sind, d.h. die müssen in einem extra track sein.

dann könnte mein importer prüfen, ob ein track monophon ist, und falls ja eine normale monophone spalte schreiben.

bevor man meinen importer mit einer midi datei füttert, lässt man sie auf monophone bzw. polyphone tracks prüfen 
(das tool sollte die anzahl gefundener überlappener noten pro Track/channel kombi ausgeben, als indiz).

und behebt tracks, von denen man weiß / will, dass sie monophon sind.

außerdem kann der importer eine option bekommen, dass bestimmte track eigenschaften übergeben werden können, wie z.B.
- track ist monophon
- track ist drumtrack / keyswitch track (keine Töne, sondern MN und staccato ::)
eigentlich reichen diese eigenschaften, so dass jeder track, der nicht polyfon ist
mit entweder der eigenschaft 'mono', 'drums' oder 'ks' versehen wird (wobei 'drums' und 'ks' zunächst das gleiche machen werden) 
also z.B. --trackconfig='violin1:mono,drum:drums,violin1ks:drums'

eine andere option des importers könnte sein, eben alle tracks mit namen auszugeben und auf überlappene noten
zu prüfen und einem das ergebnis anzuzeigen. indiz für einen drum track könnten die gm noten für kick und snare sein,
der name drums, hihat, etc., der midi kanal 10; indiz für einen keyswitch track könnten extrem niedigre noten in den untersten drei
oktaven sein.


### big import rewrite und Gruppenspalten (ermöglichen "vertikale" Imports, bzw. auslagern von Spaltengruppen)
DIE GANZE AUFLÖSEREI MIT ALLEN REGELN UND UMFANGREICHEN TESTS IST SCHON PROGRAMMIERT (reference package), DORT IN DER docs.go
STEHT AUCH, WIE MAN DIE PACKAGE BENUTZT. ES FEHLT NUR DIE ANBINDUNG AN DEN PARSER, D.H. DAS EIGENTLICHE IMPORTIEREN.
ZUM VERSTÄNDNIS RUHIG WEITERLESEN

FÜR ITEM GROUPS und ITEM GROUP modifiers (wie z.B. Tokens, Scales und Multiitems) soll das Slicing mit [] erhalten bleiben, da
dort die pattern commands nicht zu verfügung stehen und parts auch nicht benötigt werden.

- es gibt eine einheitliche syntax für imports und referenzen, d.h. es gibt nur noch referenzen
- eine referenz ist eine art pfad und bezieht sich auf einen kontext pfad in den sie importiert wird
- der pfad besteht aus (groß nach klein, außen nach innen, links nach rechts):
  - dem dateinamen, vorangestellt mit ' unter weglassen der Endung
  - dem tabellennamen, vorangestellt mit = für Scores und . für shortcuts
  - dem spaltennamen, vorangestellt mit .
  - dem partnamen, in eckigen klammern
- ein vollständiger pfad wäre z.B.
  'piano=SCORE.timbre[CHORUS]
  oder
  'UVI.ks.tremolo.cello 
- für Dateien gelten die normalen auflösungsregeln für dateien (lokal zu global)
- der kontext, in den importiert wird, wird in der programmierung zur "vervollständigung" des gewünschten pfades immer berücksichtigt und ist dort immer vollständig:
  - beim import in eine Datei aus dem vervollständigen dateinamen mit ordner
  - beim import in eine shortcut tabellenzelle aus dem dateinamen, dem tabellennamen, der zeile und der spalte
  - beim import in eine score tabelle aus dem dateinamen und dem tabellennamen
  - beim import in eine score tabellenzelle aus dem dateinamen, dem tabellennamen und der spalte
- importregeln:
  - import in eine Datei: 
    - wird zu importierende Datei komplett importiert: nur dateiname
    - wird nur eine tabelle der Datei importiert: hinter dateiname noch tabellenname
  - import in eine shortcut tabellenzelle:
    - wenn datei angegeben: aus fremder datei, sonst aus der gleichen datei
    - tabellenname immer angeben
    - zeile immer angeben
    - wenn spalte angegeben, andere spalte, sonst spalte mit gleichem namen wie zielspalte
  - import in eine score tabelle:
    - wenn datei angegeben: aus fremder datei, sonst aus der gleichen datei
    - wenn tabellen name angegeben: abweichender tabellenname, sonst gleicher name wie zieltabelle 
      (nicht bei gleicher datei möglich)
      es können nur score tabellen in score tabellen importiert werden
    - wenn partname angegeben: nur diesen part importieren, sonst gesamte tabelle
  - import in eine score tabellenzelle:
    - wenn datei angegeben: aus fremder datei, sonst aus der gleichen datei
    - wenn tabellen name angegeben: abweichender tabellenname, sonst gleicher name wie zieltabelle 
      (auch bei gleicher datei möglich), bei score Tabellen muss allerdings immer ein 
      gleichheitszeichen erscheinen, bei shortcuts folgt automatisch der punkt (über die zeile)
    - wenn Spaltenname angeben: abweichender spaltenname, sonst gleicher name wie zielspalte (nicht möglich, wenn
      import aus gleicher datei und gleichem tabellennamen), bei kombispalten werden bei weglassen des spaltennamens
      alle spalten gleichen namens importiert
    - wenn partname angegeben: nur diesen part importieren, sonst gesamte spalte
- kombispalten:
  - um es (vor allem im main score) zu ermöglichen, auch score spalten gruppen in eigene dateien oder tabellen auszulagern 
    (z.B. für controller oder keyswitches), kann man kombispalten erstellen:
    - eine kombispalte enthält mehrere tracknamen als bezeichnung im spaltenkopf diese werden durch leerzeichen getrennt, z.B.
      violin1KS violin2KS violaKS celloKS
      das wäre eine kombispalte aus den tracks violin1KS, violin2KS, violaKS, celloKS
    - in kombination mit spaltenimports lassen sich damit sehr komfortabel spaltengruppen auslagern:

```
=SCORE | violin1 violin2 viola cello | violin1KS violin2KS violaKS celloKS |
#INTRO
  1    | 'intro                      | 'keyswitch=intro                      |
*10
#A
  1    | 'movement[A]                | 'keyswitch=a                          |
*20
#B
  1    | 'movement[B]                | 'keyswitch=b                          |
*12
```

im obigen score werden die Spalten violin1, violin2, viola und cello an den entsprechenden aus den jeweiligen
spalten der =SCORE tabellen der dateien intro.mskl und movement.mskl importiert und
in eigene trackspalten verwandelt, wobei aus der score tabelle der movement.mskl nur jeweils die part [A] und [B]
übernommen werden

die spalten violin1KS, violin2KS, violaKS und celloKS kommen hingegen aus der gleichen datei keyswitch.mskl,
jedoch aus den verschiedenen tabellen  =intro, =a und =b und deren gleichlautenden spalten
    
diese ganzen auflösungsregeln werden in einem eigenen package reference ausgelagert (ist schon programmiert).

DIE IDEE MIT DEN GRUPPEN SPALTEN KANN MAN AUCH NOCH AUSBAUEN, SO DASS DINGE, DIE IN DEN GLEICHEN SPALTEN VORKOMMEN
SOLLEN, NUR EINMAL IN EINE GRUPPENSPALTE GESCHRIEBEN WERDEN MÜSSEN.


damit das alles auch mit den parts in den tabellenzellen funktioniert, muss die Slice Syntax [2:5] aufgegeben werden.
ich glaube, das ist kein großer Verlust, denn
1. können wir das gleiche, sogar verständlicher über pattern funktionen ereichen, wobei wir bei eins zu zählen anfangen, z.B.
   $slice(=mypatt,1,3)
   $from(=mypatt,3)
   $until(=mypatt,5)
2. ist die [3:6] syntax kryptisch und programmierer speak (0 ist das erste element, das erste und letzte kann weggelassen werden etc)
3. wird es nicht so oft benötigt, dass es ein eigenes zeichen rechtfertigen würde


- fix mehrtaktiges pattern, dass mitten im takt endet


patterns und wiederholungen

es gibt folgende patternsituationen:

A startposition erster Takt
B startposition erstes Event
X endPosition letzter Takt
Y endPosition durch end-Event

damit gibt es folgende Kombinationsmöglichkeiten:
AX: startposition erster Takt und endPosition ende letzten Taktes; länge ist die summe der länge der Takte
AY: startposition erster Takt und endPosition durch end-Event; länge ist position endEvent - startposition erster Takt
BX: startposition erstes Event (sync) endPosition letzter Takt; länge ist endPosition letzter Takt- startposition erstes Event
BY: startposition erstes Event (sync) endPosition durch end-Event; länge ist position endEvent - startposition erstes Event

dann gibt es folgende Wiederholungsarten:

1. Taktwiederholungen .n.
2. partwiederholungen []
3. Wiederholungen letztes event %
4. Patternwiederholung =patt%n

1. und 2. wiederholen immer die ganzen zurückliegenden takte
3. wiederholt letztes event an angegebener position
4.
    a) sync: wiederholt, indem das erste event an das ende der vorigen wiederholung anschließt
    b) kein sync: wiederholt, indem der abstand zwischen dem taktanfang und dem ersten event gleich dem abstand zwischen dem ende der vorigen wiederholung und dem ersten event der aktuellen wiederholung ist

außerdem muss man bei den patterns unterscheiden zwischen verschiedenen zuständen/phasen

phase 1: der PatternCall ist als solcher identifiziert und geparst, wir kennen name/pfad, parameter, sync, slice und wiederholungen
phase 2: die events der spalte, die zu dem pfad gehört, sind bekannt und typisiert, die länge der Spalte ist bekannt
pahse 3: die events der spalte sind parametrisiert und anhand der aufrufposition platziert und ggf. wiederholt



- pattern wiederholungssysntax verändern: von 

    =patt.col%9(a,c')[2:5]
    
  zu
  
    =patt.col(a,c')[2:5]%9

- _ beendet ein Pattern und zwar auch mitten im Takt, sodass wiederholungen auf dem _ einsetzen, d.h. man kann auch Pattern machen,
  die kürzer sind, als ein Takt!

- wir brauchen bei den tracks einen ambitus: from: to: type: (ignore,fit,mirror)
  ignore entfernt noten, die nicht reinpassen, fit verwendet die nächste passende note der skala, mirror spiegelt die note
  octaven nach unten/oben (solange, bis es passt)

- bei scale notes sind " und ' als oktavierungszeichen aktiv. d.h. wenn die scalenote > 0 wird aufwärts oktaviert und wenn sie < 0, wird abwärts oktaviert

- make a scale out of a multiitem:

    =SCORE    | p |
    # \(c e g)  // below comes the g
      1       | 2 |

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
