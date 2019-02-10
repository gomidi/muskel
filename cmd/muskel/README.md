
# MuSkeL - Musical Sketch Language

The Musical Sketch Language (MuSkeL) is an attempt to improve and unify
the Western Standard Notation in a way, that makes it easier to learn,
read, write, type and understand. There is no need for lined paper or 
special fonts, because all that is used is well known ASCII code.

## Specification

0. There must be a = before the first system. Everything above is considered header (temperament, pattern definitions)
1. The timeline (so the bars) is from top to bottom and the instruments are in columns.
   The first column contains the time positions. The following columns contain the
   event for an instrument in that time position. The last column contains naming of
   parts and jumps to parts.
   An empty line denotes a new bar. optionally there might be a time signature at the beginning of the
   line.
   the instruments should be defined in the first row where the first column is empty and the following columns
   contain the name of the corresponding instrument.
   then may follow a line for MIDI channel numbers (first column contains "Ch"), and MIDI Program numbers 
   (first column contains "Prog").
   Then come the bars and events for the instruments.
   The following events are possible:
   - notes, e.g. a'
   - drumnotes e.g. sn
   - patterns e.g. $a
   - Lyrics e.g. "Hi" lyrics must not contain whitespace
   - MIDI controller events e.g. PB200
   - rest e.g. _
   - overbind the last event from the last bar e.g. =
   - samples with a timespan e.g. Sa
   - samples with no timespan ("oneshot") e.g. Za
   - repitition of last event e.g. %
   - multiitems, consisting of several items happening at the same time. e.g.  a' PB200
     They are simply separated by whitespace.
   An event that has a timespan (like notes, motivs, rests and samples with a timespan) ends when the next event that has an timespan happens (or the   
   next bar appears, if that does not overbind).
   This results in monophonic events or (in the case of multievents) parallel events.
2. Columns are separeted with the pipe symbol `|`.
4. A multiitem consists of multiple events separated by whitespace.
   examples
   a' PB200
   e "La"
   c' e' g'
   Sa a'
   The following events may be inside a multiitem:
   notes, drumnotes,motivs,MIDI controller events, samples, patterns, lyrics
   The may be combined freely, although it is recommended to not combine notes and drumnotes. There should not be more than one lyrics item at a time
5. Rests are denoted by the underscore symbol _ . If there is no note on position 1,a bar always starts 
   with an implicit rest that does not have to be denoted. The exception is if the bar starts with the equals 
   sign = then the last timespan item of the previous bar note keeps playing until the next event. 
6. The time position is denoted by the number of the quarternote in the bar and can be
   followed by furher segmentation. E.g. in a 4/4 bar there would be the quarter note positions as counted: 1, 2, 3 and 4.
   These would denote the position on these exact quarternotes.
   Positions in between the quarternotes (counting "and") are denote by the previous quarternote followed by the & sign.
   E.g.  3& is the position between the 3rd and the 4th quarternote (counting "three and").
   Positions that are a between the "ands" and the numbers are denoted by the decimal point . after the previous position.
   E.g.  3. is the position between 3 and 3&, while 3&. is the position between 3& and 4.
   A further division into 32th will be indicated by the semicolon ;
   E.g.  3;   3.;  3&; 3&.; 
7. The following basic notes are available (from low to high)
   C" D" E" F" G" A" B" C' D' E' F' G' A' B' C D E F G A B 
   c d e f g a b c' d' e' f' g' a' b' c" d" e" f" g" a" b" c"' d"' e"' f"' g"' a"' b"'
   Each basic note may have one of the augmenting symbols # ^ °. If not otherwise defined, # means an addition
   of 50cents, ° means an addition of 25cents, and ^ means an addition of 75cents.
   The augmenting symbol is written after the letter, but before the ' or ".
   If not otherwise defined, a' is 440Hz, and there is equal temperament.
8. The position sign to the left can be omitted, if the subdivision refers to the previous position
   E.g. 
   3
   & 
   is equivalent to 
   3 
   3&

   3& 
   . 
   is equivalent to 
   3& 
   3&.

   3&. 
   ;
   is equivalent to 
   3&.
   3&.;
10. The last item is repeated by the % sign
   e.g.
   1 | a#"| 
   2 | %  |
   is the same as
   1 | a#"| 
   2 | a#"|
   
9. A division on 3 or some other number can be achieved by denoting the position
   followed by the curly baces
   E.g.
   1 | {c,d,e}2 |
   Here we have 3-tuple on position 1.
   The events appear inside the curly braces, separated by comma.
   After the closing brace comes the time position where n-tuple ends.
   The n of the n-tuple is determined by the number of events inside the curly braces
   a "hold" entry can be assigned by the equals sign =
   e.g.
   1 | {c,=,e}2 |
   here the c is held for two times of the triole
10. For drums there are special 2-letter combinations serving as pseudonotes
	kd Kick-Drum
   	sn Snare-Sound
   	ho Hihat Open
   	hc Hihat Closed
   	rd Ride
   	tb tambourine
   	tl tom low
   	tm tom middle
   	th tom high
   	sh crash cymbal
	e.g.
	4/4
	1 | kd sh |
	 &| hc    |
	2 | sn hc |
	 &| kd ho |
	3 | kd hc | 
	4 | sn hc |
	 &| ho    |
	however it makes more often sense to have separate lines, e.g.
	
	  | KkSn | Cym         |
	4/4
	1 | kd   | sh          | 
	 &|      | hc          |  
	2 | sn   | hc          |    
	 &| kd   | ho          |    
	3 | kd   | {hc,hc,hc}4 |    
	 &|      |             |  
	4 | sn   |             |
	 &|      | ho          |
	
	
11. Sx is reserved for a sample on key x (x is a note) with a timespan (plays until next timespanned item appears)
	Zx is reserved for a sample on key x (x is a note) that will play until the end ("oneshot")
	$x is reserved for patterns, where x is one or more letters [a-z][A-Z], then may follow any combination of single and double quotes
	Ox is reserved for osc receivers where x is one lowercase letter [a-z]
	
	MIDI controll messages (won't stop other items)
    The following two uppercase letters have special meaning (followed by an integer value)
	PBn  MIDI Pitchbend message where n is -8192 to 8192
	ATn  MIDI aftertouch message where n is 0 to 127
	PTxn  MIDI polyphonic aftertouch message, where x is a note (the key) and n is 0 to 127
	CCxn  MIDI CC message, where x is a letter a to z refering to a CC number (assoc defined in preamble) and n is a value 0 to 127
	these special MIDI control messages have to be in their own line that starts with Cn where n ist the MIDI channel 1-16
	There may also be areas of sliding, e.g.
	<PB500{    } 
	here Pitchbend messages are generated upwards to the target of 500 (starting from the last Pitchbend message)
	>PTa"10
	here polyphonic aftertouch messages on key a" are generated downwards to the target of 10 (starting from the last polyphonic aftertouch messages on key a") 
	<xPB500{ } and >xPTa"10{ } do it exponential (²) 
12. Glissando is denoted by one tilde sign ~ after a single note item
    the glissando goes to the next event that must be a single note item. there must not be any item even no
	MIDI controls in between (otherwise it is very hard to calculate). also both events must not have multiitems 
	if an x follows the tilde, the glissando is exponential
	Glissando means that the following note is not played, but glided to. in terms of MIDI it will be simulated
	by a series of pitchbend messages. First the pitchbend range is set (not all MIDI instruments support this)
	and then the pitchbend messages are generated on a 32th timebase.
    E.g.
	2 |a~ |
	 &|d' |
	This is a linear glissando starting on position "two" on an a ending on position "two and" on d'
	2 |a~x |
	 &|d'  |
	the same, but exponential (²)
13. Dynamic of a note is written after the affected note. 
    Dynamic signs count until the next dynamic sign comes. Dynamic signs can be written while a note
	is still playing to indicate that the loudness of the note changed. At the beginning if not otherwise indicated
	it starts with normal loudness
	+    = mf
    ++   = f
    +++  = ff
    ++++ = fff
    -    = mp
    --   = p
    ---  = pp
    ---- = ppp
	+-   = sfz
	*    = reset to normal loudness
	E.g.
	3 | a+    |
	& | c'+++ |
	On "three" the a in mezzoforte, on "three and" the c' in fortissimo  
14. Time signature changes are indicated with a/b at the beginning the empty bar line where a is the numerator and b is the denominator
    6/8 
	is a 6 8ths time signature. If no time signature is written, it is 4/4 (default)     
15. (linear) crescendo is denoted as <{  } and (linear) descrescendo as >{  } where { marks the beginning and } marks the end.
    It can be written either in a separate column (than its for all instruments) or around the affected notes
	E.g.
	4/4
	1 |d      |
	2 |<{a--- | 
	3 |c'     |
	& |g'++}  |
	
	1 |>{c"+++|
	
	1 |=--}   |
	Here we have a normal loud d on "one", followed by a pianissimo a on "two" which is part of a crescendo
	that culminates in the fortissimo c" on "one" which decrescendos to pianissimo again.
	exponential (²) crescendo and decrescendo is denoted as <x{} and >x{}
	crescendo and decrescendo do not affect notes that have assigned dynamics, but the notes in between.
	however hold notes can get a crescendo / descrescendo (in MIDI via breathcontrol), when denoted like this
	<={} or >={}
16. accelerando is denoted as ->130{ } and decelerando as <-40{ } where the number ist target tempo. otherwise similar to crescendo and decrescendo
    tempo changes and accel/decel are noted in a seperate column named Tempo.
	A tempo change that comes with a bar change may be denoted as @BPM where the BPM is the quarternotes per minutes
	@120
	this would be a bar change with a tempochange
	3/4@90
	this would be a bar change with a timesignature change and a tempochange
19. Lyrics can be written in a multievent with the note where its sung. 
    A lyric must not include whitespace and must start and end with ".  
	A syllable can be stretched across several notes via the minus sign "-"
	e.g.
	4/4
	1  | a"  "Hu" |
	 & | c"   "-" |
	2& | g" "-ah" |
20. Chords can be written on a comment line right befor the line where they should be played

    E.g.
	4/4
	// Amaj7
	1 |      |
	// C6/9    
	3 | d    |
	
	// C6/9
	1 |      |
	2 |      |
	// C5/3
	& |      |
	
	alternatively they could be written in a seperate row called Chords and be put in " like lyrics
	and the MIDI channel could be set to "-", so that they are not played 
21. Lines starting with a // are ignored (comments)
22. Multiline comments start with /* and end with */
24. Parts are written in the last column and named by capital letters.
    jumps can only go to parts and are indicated by -> and the uppercase letter of the part. only the first repetition jumps, the second
	continues.
	E.g.
	:     A                                   
	4/4
	1   |a| A
	2   |b|
	3   |c|
	
	2&  |d| 
	3&  |e|
	4&  |d| ->A
	
	1   |a|
	
	1   |a| B
	2   |b| 
	3   |c|
	
	2&  |d| 
	3&  |e|
	4&  |d|
	
	1   |a|
	
	multiple jumps can be indicated via parenthesis prefixed by the number of repetitions
	e.g.
	2(->A)
	
26. ./. means: repeat the content of the last bar
27. ... means: keep repeating the last bar until the next not empty bar.
28. .n. means: keep repeating the last n bars until the next not empty bar where n is a number between 2 and 9
28. A pattern definition line starts with an $ some letters [a-zA-Z] and followed by any optioal combination of
    single and double quotes and a colon
    Then follows the definition of the pattern in one or more bars (bar separated by pipe symbols)
	patterns can be defined based on other patterns that have been defined in pattern lines above.
	e.g.
	$a: 1a 2b 3c | 1e 2f 3_ 4e
	$b: 2&g 4&g
	$c: $a | $b | $b | 1e 2&f
	It also can be modified by slashes after the identifier and overwrite within this slashes either
	the position or the note or both or keep the note at the position by placing a colon : inside
	e.g.
	$a': $a/: c : _ 2&g 3& :/
	would result in
	$a': 1a 2c 3c | 1_ 2&g 3&_ 4e
	By using square brackets after the pattern referer, only parts of a pattern may be referred to, like with slices in
	Go:   $a[1:]  starts from the second item of pattern $a to the end
	      $a[:3]  starts from the first item to the third
		  $a[1:3] starts from the second item to the third
	e.g.
	$f: $a[1:5]/: : g :/
	would result in
	$f: 2b 3c 1g 2f
	
	patterns may be merged by writing them as a multiitem, e.g.
	$e: ($a,$b)
	would result in
	$e: 1a 2b 2&g 3c 4&g | 1e 2f 3_ 4e
	also MIDI controls may be part of a motiv
	e.g.
	$x: 1PB34  2&PB100
	$y: 1a   2b
	$z: 1($x,$y)
	
	inside the note line they can be referred to the same way but need the position before the pattern identifier
	the positions inside the pattern will be layed out relative to the starting position of the pattern inside the note
	line. if the pattern does not fit into the bar and the following bar does not start with an =, the pattern would be
	interrupted/cut. the same happens if any other event appears before the end of the pattern.
	
	e.g.
	4/4
	1 |$b                |
	 &|$b/c' :/          |
	
	1 |=  |
	 &|$a[1:]/: 2b : : :/|
	would result in
	4/4
	2& |g | 
	4& |g |
	3  |c'|
	
	1  |g | 
	 & |a |
	2  |b |
	3& |c |
	
	if a backslash \ appears before the motiv identifier, the position of the first item in the motiv is synced to the position
	before the motiv identifier and every other position of the motiv items follows from that
	e.g.
	4/4
	1& |\$b |
	would result in
	4/4 
	1& |g | 
	3& |g |
	
	motivs/patterns may receive parameters if they are invoked, like a function
	$g(3,4)
	here the pattern $g is given the parameters 3 and 4
	in the definition of the pattern these parameters may be referred to by position
	#1 for the first parameter, #2 for the second parameter etc.
	e.g.
	$g: 1PB#1  2&PB#2
	when invoked like this
	4/4
	1 |$g(34,100) |
	this would result in
	4/4 
	1 |PB34 |  
	2&|PB100|
	
	also can patterns invoke patterns
	$h: $g(#1,40)
	
	when invoked like this
	4/4 
	1 |$h(50)[1:] |
	this would result in
	4/4
	2& |PB50|
29. Osc commands may be used like other items. First an O variable has to be defined (O followed by a lowercase letter)
    to the receiver of the OSC message (ipaddress:port)
	e.g.
	Oa: 127.0.0.1:5680
	then it can be used:
	4/4 
	2& |Oa[34,46]/play/note |  	
	here the parameters 34 and 46 are send to the path /play/note
	
	Osc messages may also be used in patterns
	
30. Random values can be generated by the question mark ?
    there are two modes for this to happen:
	if parenthesis follows after the ? then inside parenthesis there are the possible values, seperated
	by commas, from which a random one may be chosen
	e.g.
	4/4 
	2& | ?(sn,++tl,-th,_) |
	
	1  | ./.              |
	plays either a snare or a low tom or a high tom or nothing
	in the second bar it chooses again another random value
	
	the other mode is with brackets following the ?
	then inside the brackets there is the probability (in percent)
	to which the following event occurs
	e.g.
	4/4 
	2& | ?[33]sn | 
	
	1  | ./.     |
	here the snare is hit with 33 percent probability
	in the second bar there is another 33 percent chance
	 

