Ein `Sketch` beschreibt einen Teil oder ein vollständiges musikalisches Stück. Es ist eine Tabelle,
deren erste Zeile mit dem Namen des `Sketches`beginnt. Die Namen der Sketche beginnen immer mit einem 
Gleichheitszeichen `=`.
Es kann in einem Musikstück mehrere `Sketche`geben, aber nur einen Hauptsketch, der immer `=SCORE` (englisch
für Notation) heisst.
Nach dem Namen des Sketches folgen die Namen der Spalten. Beim `=SCORE`-Sketch sind dies die Spuren/Stimmen. 
In den weiteren Zeilen folgen die Zeitpositionen im Takt mit den Noten. Taktwechsel werden durch eine Raute
angezeigt `#` und die Zeit läuft von oben nach unten:

```
=SCORE  | Gesang | Flöte |
#
 1      | c      |       |
 1&     |        | e"    |
 2      | d      | *     |
 2&     | e      | f"    |
 3      |        | g"    |
 4&     | d      |       |
#
 1      | c      | c"'   |
#
 1      | *      | *     |
```

Hier haben wir zwei Stimmen bzw. Spuren (engl.: Tracks): "Gesang" und "Flöte". Wir haben drei Takte, deren Start mit 
der Raute `#` angezeigt wird. innerhalb der Takte haben wir die Zeitpositionen, die von oben nach unten verlaufen.
Eine Zeitposition wird angegeben durch den Schlag (engl. Beat), also die Viertelposition im Takt ggf. gefolgt von kleineren
Zeiteinheiten. Hierzu werden die gewöhnlichen Zahlen verwendet: 1 ist die Position des ersten Viertels, 2 die des zweiten
Viertels usw. Wer schon einmal Viertel im Takt gezählt hat, kennt das Prinzip. Genau genommen fällt die Position des ersten
Viertels mit dem Start des Taktes zusammen. Das zweite Viertel startet eine Viertelnote später, usw.
Mit jedem neuen Takt wird neu gezählt. Im Gegensatz zur westlichen Standardnotation werden nur die Startpositionen
von Noten, Pausen usw. notiert, nicht die Dauern. Jede Note dauert so lange, bis entweder die nächste Note
kommt, oder eine Pause.

Zum Beispiel dauert das kleine c im zweiten Takt in der Gesangsspur genau einen Takt - nämlich bis zur ersten Viertelposition
im dritten Takt, wo eine Pause notiert ist (durch das Sternchen `*`). 

Wir haben hier noch feinere Zeitpositionen: die zwischen den Viertelpositionen (Offbeat, Achtelnote Abstand). 
Die Zeitposition zwischen dem ersten Viertel und dem zweiten Viertel ist `1&` (gesprochen "eins und"), 
zwischen dem zweiten und dem dritten Viertel `2&`("zwei und") usw. 
Die Zeitposition, die am Anfang einer Zeile notiert ist, gilt für die gesamte Zeile.
Hier wird also im ersten Takt auf der Zeitposition `2&` gleichzeitig in der Gesangsspur das kleine e gesungen, während
die Flöte das zweigestrichene f spielt.

Die Zeitpositionen, an denen nichts stattfindet, müssen auch nicht notiert werden. So sind für den zweiten und dritten
Takt nur der erste Schlag notiert. Dafür muss jedoch die Taktdauer bekannt sein.

Wenn der Taktwechsel nur mit einer Raute notiert ist `#`, dann ist die Taktdauer genauso lang, wie diejenige des Taktes
zuvor. Falls die Taktdauer sich ändern soll, schreibt man hinter die Raute mit etwas Abstand den Zähler und Nenner des Taktes.
Diese werden durch einen Querbalken abgetrennt. Also z.B. `#  3/4` für einen Takt mit der Länge von drei Viertelnoten.

Wenn der erste Takt keinen Taktwechsel angezeigt hat, so wird der Vier-Viertel-Takt `4/4` gewählt.

Tempowechsel werden auch in die Taktzeile hinter die Raute oder den Taktdauerwechsel geschrieben. Hierbei wird die Anzahl der Schläge
je Minute (BPM: Beats Per Minute) angegeben, mit einem vorangestelltem Klammeraffen `@`. Z.B. beschreibt `# @110` einen Tempowechsel
auf 110 BPM ohne Wechsel der Taktdauer, während sowohl `# 3/8 @90` als auch `# @90 3/8` einen Wechsel der Taktdauer auf drei Achtelnoten
und einen Tempowechsel auf 90 BPM anzeigen. Der erste Takt hat, wenn er ohne Tempowechsel notiert ist, ein Tempo von 120 BPM.

Die Spalten können bei handschriftlicher Notation einfach durch durchgezogene Linien getrennt werden. Bei der Notation am Computer
wird in jeder Zeile außer den Taktwechselzeilen das Pipe-Zeichen `|` verwendet. Hierbei müssen die Pipe-Zeichen der verschiedenen
Zeilen nicht exakt untereinander stehen: das muskel Hilfsprogram kann eine Datei so formatieren, dass sie anschließend untereinander
stehen und kann komfortabel in Ihren Editor integriert werden, so dass z.B. unmittelbar nach dem Speichern die Pipe-Zeichen automatisch
"geradegerückt"
