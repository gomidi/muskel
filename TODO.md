# nächste TODOs

- allow Ab Db ebb etc.
- some improvements to make it faster to write:
 4. check that position goes not beyond bar length

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

- dokumentation aktualisieren
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
