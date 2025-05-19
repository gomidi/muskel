
Der folgende Score resultiert beim unrolling

| TRACK      | Violin |
| ---------- | ------ |
|  Channel   | 2      |
|  Transpose | 0      |


| =motiv   | violin |
| -------- | ------ |
|  #       |        |
|     1    | c      |
|     1&   |        |
|     2    | c      |
|     2&   | d      |
|     3    | d      |
|     3&   | c      |
|  #       |        |
|     1    |        |
|     1&   | c      |
|     2&   | d      |
|     3    | c      |
|     3&   | f      |


| =SCORE   | Violin        |
| -------- | ------------- |
|  # 3/4   |               |
|     1    |               |
|     1&   |               |
|     2    |               |
|     2&   | =motiv.violin |
|     3    |               |
|     3&   |               |
|  #       |               |
|     1    | ...           |
|     1&   |               |
|     2    |               |
|     2&   |               |
|     3    |               |
|     3&   |               |
|  #       |               |
|     1    |               |
|  *2      |               |
|  #       |               |
|     1    | *             |


###################

unrolled:

BUG: siehe doppelte 1 in takt 4. es scheint abhängig von der patternlänge von pattern =motiv zu sein,
aber auch abhängig von der anzahl der takte vor der wiederholung. sehr strange.

wenn das pattern mit % wiederholt wird, tritt der bug auch nicht auf.


| TRACK           | Violin                                   |
| --------------- | ---------------------------------------- |
|  File           |                                          |
|  Channel        | 2                                        |
|  Program        |                                          |
|  Bank           |                                          |
|  Transpose      |                                          |
|  Volume         |                                          |
|  Delay          |                                          |
|  PitchbendRange | 2                                        |
|  VelocityScale  | min:1 max:127 random:4 step:15 center:63 |
|  Ambitus        |                                          |
|  Import         |                                          |

| =SCORE                              | Violin |
| ----------------------------------- | ------ |
|  # 3/4 @120.00 \major^c'           //  #1|        |
|     2&                              | c      |
|     3&                              | c      |
|  #           //  #2                 |        |
|     2&                              | c      |
|     3&                              | c      |
|  #           //  #3                 |        |
|     1                               | d      |
|     1&                              | d      |
|     2                               | c      |
|  #           //  #4                 |        |
|     1                               | c      |
|     1                               | d      |
|     1&                              | d      |
|     2                               | d      |
|     2                               | c      |
|     2&                              | c      |
|     3                               | f      |
|  #           //  #5                 |        |
|     1                               | c      |
|     2                               | d      |
|     2&                              | c      |
|     3                               | f      |
|  #           //  #6                 |        |
|     1                               | *      |


