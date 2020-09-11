/*
Usage

0. Understand the meaning of the reference types:

 reference type | is context | complete example       | meaning
----------------------------------------------------------------------------------------
 FileCtx        | yes        |                        | place at the toplevel of a muskel file
 ShortCutCtx    | yes        |                        | place in a shortcut table
 ScoreCtx       | yes        |                        | place at the beginning of a line in a score table
 ScoreColCtx    | yes        |                        | place in a cell of a score table (not the header or first col)
 File           | no         | 'drums                 | import all tables from the file drums.mskl (only within FileCtx)
 ShortCut       | no         | 'drums.cc              | import the shortcut table .cc from drums.mskl (only within FileCtx)
 Score          | no         | 'drums=hh              | import the score table =hh from drums.mskl into FileCtx or ScoreCtx
 ScorePart      | no         | 'drums=hh[CHORUS]      | import the part CHORUS of the score table =hh from drums.mskl into ScoreCtx
 ScoreCol       | no         | 'drums=hh.loud         | import the column loud of the score table =hh from drums.mskl into ScoreColCtx
 ScoreColPart   | no         | 'drums=hh.loud[CHORUS] | import the part CHORUS in the column loud of the score table =hh from drums.mskl into ScoreColCtx
 ShortCutCell   | no         | 'drums.cc.timbre.hard  | import the column hard in the row timbre of the shortcut table .cc from drums.mskl into ScoreColCtx or ShortCutCtx

Since some parts of the complete syntax might be missing and be filled via the given context, the purpose
of this library is to make the completion for you.

within the parser....

1. prepare and store the current context by calling one of NewFileCtx, NewShortCutCtx, NewScoreColCtx or NewScoreCtx
2. check if the given string is a reference by calling IsReference
3. if you have a reference, parse it via Parse(ref, ctx)
4. act according to the reference type, the following reference types are available in the corresponding contexts

 context     | reference type  | what to do
--------------------------------------------------------------------------
 FileCtx     | File            | import all tables of file
             | Score           | import score table of file
             | ShortCut        | import shortcut table of file
--------------------------------------------------------------------------
 ScoreCtx    | Score           | import score into score
             | ScorePart       | import part of score into score
--------------------------------------------------------------------------
 ScoreColCtx | ScoreCol        | import score col(s) into score col(s)
             | ScoreColPart    | import part of score col(s) into score col(s)
             | ShortCutCell    | import shortcut cell into score col cell
--------------------------------------------------------------------------
 ShortCutCtx | ShortCutCell    | import shortcut cell into shortcut cell

Since the current column in ScoreColCtx could be a column group (seperated by empty space),
if the ScoreCol or ScoreColPart had no columnname given, the reference will end up with the
same column group and each column of the group must be imported, if there exists a column of the
same name inside the referenced score.

5. the complete File, Dir, Table, Row, Col and Part infos are given within the parsed reference.
The Cols() method returns any columns in a slice.

*/

package reference
