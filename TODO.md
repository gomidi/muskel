# nächste TODOs

- fix: sicherstellen, dass bei verkürzter schreibweise nur die letzte ziffer/beat position verwendet wird und nicht die letzte position, also
  4
    .   => 4.
   &    => 4&

  3&
    .  => nicht zulässig
   &.  => 3&. 
- fix: replace/overwrite nach taktwiederholungen mit .2. etc.
- fix: glissandi führen zu positionsverschiebungen
- fix: microverschiebungen zur gleichen zeit werden nicht korrekt aufgehoben (es verschiebt sich alles nachfolgende)
  2  | #1<        |
  2  | #2<        |              
  2  | #3<        |
  3  | (#1 #2 #3) |              
// oder
  2  | (#1< #2< #3<) |              
  3  | (#1 #2 #3) |

- add $embed() command with params like drums.* (would allow all tokens in the drum table to be addressed without the drums. prefix)
  or drums.sn (would only allow sn to be addressed without the drums. prefix). Multiple of these params could be passed to $embed.
  multiple calls of embed could be done. it would be great, if these "globals" would only be globals for the current file/score.
  but that would require the token resolving procedure to change a bit, I guess. (maybe its time to separate token resolving from pattern resolving, which also might make sense, since token resolving must not be recursive while pattern resolving may be)

- tempowechsel und globale scalewechsel mitten im takt ermöglichen:
  in der positionsspalte hinter der position, durch leerzeichen abgetrennt, z.B.
  2& @220 \major^g#' diese müssen im takt gespeichert werden (mit ihrer position)

- allow to use parts like patterns inside cols (via [part] syntax): 
  that allows to selectively repeat parts (just for some cols) and to override not fitting things
    
- cleanup scalenotes: make nt.ScaleNote a *int8 and copy it when duplicating. when it is not set, it is no scale note. otherwise it
  is the step (counting from 0). parsing makes 1->0 2->1 etc but -1->-1 this should make scale calculations and transpositions easier.

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
