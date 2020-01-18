package predefined

var SketchesAndTokens = map[string]string{
	"drumnote": `
// some shortcuts for drums (GM)

drums
// agogo hi
ah    | MN67:::
// agogo low
al    | MN68:::
// kick drum
kd    | MN36:::
// bongos hi
bh    | MN60:::
// bongos low
bl    | MN61:::
// clap
cl    | MN39:::
// crash cymbal
cr    | MN49:::
// hihat closed
hh    | MN42:::
// hihat open
ho    | MN46:::
// ride cymbal
rd    | MN51:::
// snare drum
sn    | MN40:::
// splash cymbal
sp    | MN55:::
// tom low
tl    | MN45:::
// tom middle
tm    | MN48:::
// tom hi
th    | MN50:::
// tambourin
tb    | MN54:::
// triangle muted / closed
tc    | MN80:::
// trianlge open
to    | MN81:::
// high conga
hc    | MN63:::
// low conga
lc    | MN64:::
// side stick
st    | MN37:::

 `,
	"chord": `

chord
I    | *^1*^3*^5*
I7   | *^1*^3*^5*^7*
II   | *^2*^4*^6*
II7  | *^2*^4*^6*^7*
III  | *^3*^5*^7*
III7 | *^3*^5*^7*^9*
IV   | *^4*^6*^8*
IV7  | *^4*^6*^8*^10*
V    | *^5*^7*^9*
V7   | *^5*^7*^9*^11*
VI   | *^6*^8*^10*
VI7  | *^6*^8*^10*^12*
VII  | *^7*^9*^11*
VII7 | *^7*^9*^11*^13*
T    | I
D    | V
S    | IV
DP   | III
SP   | II
DG   | VII
SG   | VI
TP   | VI
TG   | III

`,
	"learn": `
=learn  | melody                                          | pattern | names                                                                     |
# 4/4
1       | =learn.pattern(g#,b,c',a,a#,f#,g,f,d,d#,c#,e,c) | #1      | =learn.pattern("g#","b","c'","a","a#","f#","g","f","d","d#","c#","e","c") |
1&      |                                                 | #2      |                                                                           |
2       |                                                 | #3      |                                                                           |
3       |                                                 | #4      |                                                                           |
3.      |                                                 | #5      |                                                                           |
3&      |                                                 | #6      |                                                                           |
4       |                                                 | #7      |                                                                           |
#
1       |                                                 | #8      |                                                                           |
1.      |                                                 | #9      |                                                                           |
1&      |                                                 | #10     |                                                                           |
2&;     |                                                 | *       |                                                                           |
2&.     |                                                 | #11     |                                                                           |
3&.     |                                                 | #12     |                                                                           |
4       |                                                 | #13     |                                                                           |

`,
	"scale": `
// some shortcuts for testing scales
=scale | up  | down |
# 4/4
1      | ^1  | ^16  |
1&     | ^2  | ^15  |
2      | ^3  | ^14  |
2&     | ^4  | ^13  | 
3      | ^5  | ^12  | 
3&     | ^6  | ^11  | 
4      | ^7  | ^10  | 
4&     | ^8  | ^9   |
#
1      | ^9  | ^8   | 
1&     | ^10 | ^7   |
2      | ^11 | ^6   |
2&     | ^12 | ^5   | 
3      | ^13 | ^4   |
3&     | ^14 | ^3   |
4      | ^15 | ^2   |
4&     | ^16 | ^1   |
`,
	"repeat": `
=repeat | 8  | 16 | 32 | 3x4        | 3x8        | 3x16       | 5x4              | 5x8              | 5x16             | 7x4                    | 7x8                    | 7x16                   |
# 4/4
1       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
1;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
1&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
2;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
2&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        | 
2&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |                
3       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
3;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
3&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
4;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
4&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
#
1       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
1;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
1&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
2;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
2&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        | 
2&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |                
3       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
3;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
3&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
4;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
4&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
#
1       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
1;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
1&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
1&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
2;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
2&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
2&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        | 
2&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |                
3       | #1 | #1 | #1 | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
3;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
3&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
3&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4       | #1 | #1 | #1 |            | {#1,#1,#1} | {#1,#1,#1} |                  | {#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1} |                        | {#1,#1,#1,#1,#1,#1,#1} | {#1,#1,#1,#1,#1,#1,#1} |
4;      |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.      |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4.;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&      | #1 | #1 | #1 |            |            | {#1,#1,#1} |                  |                  | {#1,#1,#1,#1,#1} |                        |                        | {#1,#1,#1,#1,#1,#1,#1} |
4&;     |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.     |    | #1 | #1 |            |            |            |                  |                  |                  |                        |                        |                        |
4&.;    |    |    | #1 |            |            |            |                  |                  |                  |                        |                        |                        |

`,
	"cc": `
// some shortcuts for controllers
cc              |
 bank           | CC(0,#1)
 modulation     | CC(1,#1)
 mod            | CC(1,#1)
 modwheel       | CC(1,#1)
 mw             | CC(1,#1)
 breath         | CC(2,#1)
 foot           | CC(4,#1)
 portamento     | CC(5,#1)
 porta          | CC(5,#1)
 data           | CC(6,#1)
 volume         | CC(7,#1)
 vol            | CC(7,#1)
 balance        | CC(8,#1)
 bal            | CC(8,#1)
 pan            | CC(10,#1)
 expression     | CC(11,#1)
 exp            | CC(11,#1)
 effect1        | CC(12,#1)
 fx1            | CC(12,#1)
 effect2        | CC(13,#1)
 fx2            | CC(13,#1)
 slider1        | CC(16,#1)
 slider2        | CC(17,#1)
 slider3        | CC(18,#1)
 slider4        | CC(19,#1)
 sl1            | CC(16,#1)
 sl2            | CC(17,#1)
 sl3            | CC(18,#1)
 sl4            | CC(19,#1)
 bank.lsb       | CC(32,#1)
 modulation.lsb | CC(33,#1)
 mod.lsb        | CC(33,#1)
 modwheel.lsb   | CC(33,#1)
 mw.lsb         | CC(33,#1)
 breath.lsb     | CC(34,#1)
 foot.lsb       | CC(36,#1)
 portamento.lsb | CC(37,#1)
 porta.lsb      | CC(37,#1)
 data.lsb       | CC(38,#1)
 volume.lsb     | CC(39,#1)
 vol.lsb        | CC(39,#1)
 balance.lsb    | CC(40,#1)
 bal.lsb        | CC(40,#1)
 pan.lsb        | CC(42,#1)
 expression.lsb | CC(43,#1)
 exp.lsb        | CC(43,#1)
 effect1.lsb    | CC(44,#1)
 fx1.lsb        | CC(44,#1)
 effect2.lsb    | CC(45,#1)
 fx2.lsb        | CC(45,#1)
 variation      | CC(70,#1)
 var            | CC(70,#1)
 timbre         | CC(71,#1)
 tim            | CC(71,#1)
 release        | CC(72,#1)
 rel            | CC(72,#1)
 attack         | CC(73,#1)
 att            | CC(73,#1)
 brightness     | CC(74,#1)
 bright         | CC(74,#1)
 control6       | CC(75,#1)
 control7       | CC(76,#1)
 control8       | CC(77,#1)
 control9       | CC(78,#1)
 control10      | CC(79,#1)
 ctrl6          | CC(75,#1)
 ctrl7          | CC(76,#1)
 ctrl8          | CC(77,#1)
 ctrl9          | CC(78,#1)
 ctrl10         | CC(79,#1)
 reverb         | CC(91,#1)
 rev            | CC(91,#1)
 effects        | CC(91,#1)
 fx             | CC(91,#1)
 tremulo        | CC(92,#1)
 trem           | CC(92,#1)
 chorus         | CC(93,#1)
 chor           | CC(93,#1)
 celeste        | CC(94,#1)
 cel            | CC(94,#1)
 phaser         | CC(95,#1)
 phs            | CC(95,#1)
 data.inc       | CC(96,#1)
 data.dec       | CC(97,#1)
 nrpn.lsb       | CC(98,#1)
 nrpn           | CC(99,#1)
 rpn.lsb        | CC(100,#1)
 rpn            | CC(101,#1)
 sound.off      | CC(120,0)
 snd.off        | CC(120,0)
 controllers.off| CC(121,0)
 ctrl.off       | CC(121,0)
 notes.off      | CC(123,0)
 omni.off       | CC(124,0)
 omni.on        | CC(125,0)
 mono.on        | CC(126,127)
 poly.on        | CC(127,0)
 local.on       | CC(122,127)
 local.off      | CC(122,0)
 hold.on        | CC(64,127)
 hold.off       | CC(64,0)
 portamento.on  | CC(65,127)
 porta.on       | CC(65,127)
 portamento.off | CC(65,0)
 porta.off      | CC(65,0)
 sustenuto.on   | CC(66,127)
 sust.on        | CC(66,127)
 sustenuto.off  | CC(66,0)
 sust.off       | CC(66,0)
 soft.on        | CC(67,127)
 soft.off       | CC(67,0)
 legato.on      | CC(68,127)
 legato.off     | CC(68,0)
 leg.on         | CC(68,127)
 leg.off        | CC(68,0)
 hold2.on       | CC(69,127)
 hold2.off      | CC(69,0)
 switch1.on     | CC(80,127)
 switch1.off    | CC(80,0)
 switch2.on     | CC(81,127)
 switch2.off    | CC(81,0)
 switch3.on     | CC(82,127)
 switch3.off    | CC(82,0)
 switch4.on     | CC(83,127)
 switch4.off    | CC(83,0)
 sw1.on         | CC(80,127)
 sw1.off        | CC(80,0)
 sw2.on         | CC(81,127)
 sw2.off        | CC(81,0)
 sw3.on         | CC(82,127)
 sw3.off        | CC(82,0)
 sw4.on         | CC(83,127)
 sw4.off        | CC(83,0)

=nrpn | nrpn            | reset            | inc             | dec             |
 #
 1    | cc.nrpn(#1)     | cc.nrpn(127)     | cc.nrpn(#1)     | cc.nrpn(#1)     |
 1;   | cc.nrpn.lsb(#2) | cc.nrpn.lsb(127) | cc.nrpn.lsb(#2) | cc.nrpn.lsb(#2) |
//    | set value msb   |
 1.   | CC(6,#3)        |                  | CC(96,0)        | CC(97,0)        |
//    | set value lsb   |
 1.;  | CC(38,#4)       |                  | =!.reset        | =!.reset        |
 1&   | =!.reset        |                  |                 |                 |

=rpn | rpn            | reset           | inc            | dec            |
 #
 1   | cc.rpn(#1)     | cc.rpn(127)     | cc.rpn(#1)     | cc.rpn(#1)     |
 1;  | cc.rpn.lsb(#2) | cc.rpn.lsb(127) | cc.rpn.lsb(#2) | cc.rpn.lsb(#2) |
//   | set value msb  |
 1.  | CC(6,#3)       |                 | CC(96,0)       | CC(97,0)       |
//   | set value lsb  |
 1.; | CC(38,#4)      |                 | =!.reset       | =!.reset       |
 1&  | =!.reset       |                 |                |                |


=tuning | fine             | coarse           | program          | bank             |
 #
  1     | =!rpn(0,1,#1,#2) | =!rpn(0,2,#1,#2) | =!rpn(0,3,#1,#2) | =!rpn(0,4,#1,#2) |

=pitchbend | range           |
 #
  1        | =rpn(0,0,#1,#2) |

=reset | reset                                            |
 #
  1    | cc.bank(#1)                                      |
  1.   | PC(#2)                                           |
  1.;  | cc.ctrl.off                                      |
  1&   | (cc.vol(100) cc.exp(127) cc.hold.off cc.pan(64)) | 
  1&.  | =!pitchbend.range(2,0)                           |
`,
}