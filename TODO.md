# nächste TODOs

- fix problem with lyrics

- allow notes and midinotes to be transposed, e.g. c'^12 is the same as c" and c'^12^-4 is g#'
  ^2^4 is the same as ^6 and ^2^4^-3 is ^3

- allow conflicting dynamics that get "added": e.g. c'+- is the same as  c' and c'+-+ will be c'+
  that allows for placeholders to simply append dynamics, e.g.  #1+ may as well be called with c'-

- tempowechsel und globale scalewechsel mitten im takt ermöglichen:
  in der positionsspalte hinter der position, durch leerzeichen abgetrennt, z.B.
  2& @220 \major^g#' diese müssen im takt gespeichert werden (mit ihrer position)

- $_euclid with tokens, patterns, multiitems, ntuples etc

- fix: glissandi führen zu positionsverschiebungen

- replace/overwrite nach taktwiederholungen mit .2. etc.

- allow to use parts like patterns inside cols (via [part] syntax): 
  that allows to selectively repeat parts (just for some cols) and to override not fitting things
    
- what happens with =pattern.x[3:10] if there is a bar in between? does it even matter?

- embed ntuples inside ntuples, e.g. {c,d,{e,B}}

- improve/fix scale definition syntax in a way that chords could also be defined as a scale
- fix mounting patterns / scale notes within patterns
- allow repeat last event sign to alter volume and scale mounting
- allow tempo changes inside columns of main score, independant from main tempo with @130 etc. to set back to the main tempo, use
  just @. also relative tempo possible, e.g. @50%

- fix tests for tokens, sketches and patterns
- fix schreiben in mehrere dateien
- dokumentation aktualisieren
- check and fix randomness
- tests reanimieren und code cleanup, Packages, proper sketch/column definition (vielleicht wird column zu sketch und die sketchtable zu system)

## weiteres / Komplexeres

- tuning
- timbres
- Ducking?
- konvertieren: SMF zu muskel

## erledigt

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
