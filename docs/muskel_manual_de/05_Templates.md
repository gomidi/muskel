# Templates

Templates sind die flexibelste und komplexeste Eigenschaft von muskel.
Sie erlauben es, schnell viele Variation eines Motivs, Themas oder rhythmichen Patterns.
zu erstellen und im Notationsbereich mit diversen Arrangements zu experimentieren.

Templates werden im Kopfbereich definiert und werden im Notationsbereich genutzt.
Templates können aber genauso selbst andere Templates nutzen und variieren - die Syntax
dazu ist die gleiche, wie jene zur Verwendung im Notationsbereich.

Daher werden wir zunächst einfache Templates definieren und dann sehen, wie sie genutzt werden können.


## Definition eines Templates

Ein Template wird im Kopfbereich definiert und besteht aus einem Namen und dem Inhalt des Templates.
Um ein Template zu definieren, schreiben wir den Namen des Templates an den Anfang der Zeile im
Kopfbereich gefolgt von einem Doppelpunkt. Nach dem Doppelpunkt folgt der eigentliche Inhalt des Templates.

Der Name eines Templates besteht aus mindestens zwei Zeichen, wobei das erste Zeichen eines der folgenden
sein muss a-z oder A-Z und alle weiteren Zeichen zusätzlich noch die Zahlen 0-9 und den Unterstrich _ beinhalten
dürfen.

Der Inhalt des Templates ist im Prinzip das, was auch in einer Spalte einer Spur stehen kann, nur das von 
links nach rechts geschrieben wird, die Taktposition vorangestellt wird, die Ereignisse per Leerzeichen abgetrennt
werden und es keine Taktarten und Taktwechsel gibt.

Klingt verwirrend? Hier ein Beispiel

```
ah_ja:  1a 2&g' 3&g' 5c
```

Hier ist der Name des Templates ah_ja und der Inhalt 1a 2&g' 3&g' 5c
Wir haben also 4 Ereignisse: auf der 1 die note a auf 2& g' auf der 3& g' und auf der 5c
Aber wir haben keinen Taktwechsel. Wenn wir dieses Template in einen 4/4 Takt einbetten, landet
das c auf der 1 des nächsten Taktes:

```
ah_ja:  1a 2&g' 3&g' 5c

=
      |piano|
4/4 
    1 |ah_ja|

    1 |     |
```

resultiert in:

```
=
      |piano|
4/4 
    1 |a    |
    2&|g'   |
    3&|g'   |

    1 |c    |
```

oder beim 3/4 Takt

```
ah_ja:  1a 2&g' 3&g' 5c

=
      |piano|
3/4 
    1 |ah_ja|

    1 |     |
```

resultiert in:

```
=
      |piano|
3/4 
    1 |a    |
    2&|g'   |
    3&|g'   |

    2 |c    |
```

Hier sehen wir, dass Templates unabhängig von der Spur (und damit dem Instrument) und
vom Takt sind.

Sie laufen solange, bis in der Spur etwas anderes kommt:

```
ah_ja:  1a 2&g' 3&g' 5c

=
      |piano|
3/4 
    1 |ah_ja|

    1 |d'    |
```

resultiert in:

```
=
      |piano|
3/4 
    1 |a    |
    2&|g'   |
    3&|g'   |

    1 |d'   |
```

Der Übersichtlichkeit halber kann man Templates auch mit "Hilfstakten" versehen.
Dazu ist die Angabe der Taktart hinter dem Template-Namen erforderlich

```
ah_ja@4/4:  1a 2&g' 3&g' | 1c
```

entspricht

```
ah_ja:  1a 2&g' 3&g' 5c 
```

Es ist immer nur ein Hilfstakt pro Template-Definition möglich.
Bei Taktwechseln also aufsplitten:

```
tempA@4/4: 1c 2d 3e | 2f 3g
tempB@3/4: 2d' | 1e'
tempCombi: 1tempA 9tempB 15tempA
```

entspricht:

```
tempCombi: 1c 2d 3e 6f 7g 10d' 12e' 15c 16d 17e 20f 21g
```

## Aufruf eine Templates

Wie wir schon gesehen haben, geschieht der Aufruf eines Templates einfach durch das Schreiben des Namens
des Templates. Allerdings gibt es dabei noch ein paar Optionen:

1. Es können mit eckigen Klammern Ausschnitte des Templates gewählt werden, z.B. wählt
   ah_ja[1:3]
   nur die noten 2&g' und 3&g'

2. Es können die Ereignisse ersetzt werden
   ah_ja/:,d',:,6/
   ersetzt das Ereignis 2&g' zu 2d' und 5c zu 6c

   Wenn Pausen beim Ersetzen übersprungen werden sollen, bekommen die Ersetzer ein `!` vorangestellt:

```
ah_ne:  1a 2&* 3f 4* 5g 6*
ah_so:  1ah_ne/!d',!g,!5&/
```

ersetzt das Ereignis 1a zu 1d', 3f zu 3g und 5g zu 5&g



3. Es können im Template Platzhalter mit #n definiert werden, wobei n eine Zahl ist und
   beim Aufruf werden diese dann ersetzt.
   z.B.

```
ah_ja:  1a 2&#1 3&#1 5#2

=
      |piano      |
3/4 
    1 |ah_ja(d',f)|

    1 |           |
```

resultiert in:

```
=
      |piano|
3/4 
    1 |a    |
    2&|d'   |
    3&|d'   |

    1 |f    |
```

4. Es kann die Dynamik verändert werden z.B.

```
ah_ja:  1a- 2&g'+ 3&g'++ 5c=

=
      |piano |
3/4 
    1 |ah_ja+|

    1 |      |
```

resultiert in:

```
=
      |piano|
3/4 
    1 |a=   |
    2&|g'++ |
    3&|g'+++|

    1 |c+   |
```

Außerdem können Templates noch auf zwei Weisen in die Zeitachse eingehängt werden und
bei Skalennoten auch in die Skala (s.u.)

Alle diese Varianten können auch kombiniert werden:

```
ah_ja:  1a- 2&#1+ 3&#1++ 5#2=

=
      |piano                   |
3/4 
    1 |ah_ja+(d',f)[1:4]/:,:,6/|

    1 |                        |
```

resultiert in:

```
=
      |piano|
3/4 
    2&|d'++ |
    3&|d'+++|

    2 |f+   |
```

Und diese Aufrufformen können auch in einem Templateinhalt stehen:

```
ah_ja:  1a- 2&#1+ 3&#1++ 5#2= 
ah_ne:  1ah_ja+(d',f)[1:4]/:,:,6/
```

### Einhängen in die Zeitachse

### Einhängen in die Skala 
# Scalen, Pattern etc

Als Alternative zu absoluten Noten mit a" usw. kann man Stufen einer Skala wählen.

```
^0 nicht definiert
^1  Nullpunkt der Skala (1.Stufe)
^2  eine Stufe aufwärts vom Nullpunkt der Skala (2. Stufe)
^-3 drei Stufen abwärts vom Nullpunkt der Skala (3. Unterstufe)
```

Die aktuelle Scala wird im Score mit einem $scale() command gesetzt.
Der erste Wert ist der absolute Referenzton (Nullpunkt) der Skala und der zweite Wert ist der Name der
Skala, z.B.

```
$scale(a", minor)
```

Ist eine Taktangabe und setzt die aktuelle Scala auf a-Moll, wobei der Nullpunkt die Note a" ist.
Scalen sind zunächst fest im Code definiert. Als Ausgangspunkt starten wir mit
den kirchentonarten, sowie abkürzungen für Dur und Moll (minor, major)

Scalen gelten immer für alle Instrumente. Sie verändern aber nicht MIDINotes (z.B. für Schlagzeug, Switch-Keys etc)
und auch nicht Notes im absoluten Modus. Notes haben zwei Modi: den ursprünglichen absoluten Notenmodus und den Scalenmodus.

Ein $scale command gilt immer bis zum nächsten.

Bei Patternersetzungen können frei absolute Noten durch Scalennoten ersetzt werden und umgekehrt.
MIDI-Transponiert werden beide. 

Scalennoten in Pattern können verschoben (^) und eingehängt (^^) werden.

## Verschiebung

z.B.

pattA: 2^3 2&^5 4^1
pattB: pattA^-2

hier wird pattA beim Aufruf in pattB um zwei Skalenschritte hochverschoben
pattB ist also de facto

pattB: 2^1 2&^2 4^-2

dies kann auch mit ersetzungen kombiniert werden:

pattB: pattA^2/:,:,^5/

entspräche

pattB: 2^5 2&^7 4^5

Beim verschieben werden nur Skalennoten verändert. absolute Noten bleiben unberührt.

## Einhängen

Mit zwei Dächern können Noten in eine Skala an einer bestimmten Position eingehängt werden.
Ist die erste Note des eingehängten Patterns eine Skalennote, so werden nur Skalennoten verändert,
ist sie eine absolute Note, so werden nur absolute Noten verändert.

Besonders interessant ist dies in Kombination mit der Parametrisierung: Wird die erste Note über einen Parameter gesetzt, kann somit
beim Aufruf zwischen den Ersetzungsvarianten gewechselt werden.

Beim einhängen wird die erste Note auf die einhäng Position gesetzt und alle weiteren Noten werden gemäß ihrer Relation
zur ersten Note verändert. Ist die erste Note eine absolute Note, werden so entstehende skalenfremde Noten auf die nächsthöhere 
Skaleneigene Note gesetzt.

die volle Syntax für einen patterncall lautet (wenn alle Optionen genutzt werden können)

?50%!name^^3++(^2,a")[3:5]/:,2,:/

Die Werte, die als Parameter übergeben werden, werden angewendet, bevor das einhängen stattfindet.
(das Slicen und Ersetzen findet aber nach dem Einhängen statt)

Einhängen von Patterns in eine Scala kann nur im Score geschehen (nicht in Pattern definitionen)


