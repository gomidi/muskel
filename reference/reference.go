package reference

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

/*
### big import rewrite
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
    - eine kombispalte enthält mehrere tracknamen als bezeichnung im spaltenkopf diese werden durch leerzeichen getrenntz.B.
      violin1KS violin2KS violaKS celloKS
      das wäre eine kombispalte aus den tracks violin1KS, violin2KS, violaKS, celloKS
    - in kombination mit spaltenimports lassen sich damit sehr komfortabel spaltengruppen auslagern:

```
=SCORE | violin1 violin2 viola cello | violin1KS violin2KS violaKS celloKS |
#INTRO
  1    | 'intro                      | 'keyswitch=intro                    |
*10
#A
  1    | 'movement[A]                | 'keyswitch=a                        |
*20
#B
  1    | 'movement[B]                | 'keyswitch=b                        |
*12
```

im obigen score werden die Spalten violin1, violin2, viola und cello an den entsprechenden aus den jeweiligen
spalten der =SCORE tabellen der dateien intro.mskl und movement.mskl importiert und
in eigene trackspalten verwandelt, wobei aus der score tabelle der movement.mskl nur jeweils die part [A] und [B]
übernommen werden

die spalten violin1KS, violin2KS, violaKS und celloKS kommen hingegen aus der gleichen datei keyswitch.mskl,
jedoch aus den verschiedenen tabellen  =intro, =a und =b und deren gleichlautenden spalten


diese ganzen auflösungsregeln werden in einem eigenen package reference ausgelagert.
*/

type Type int

const (
	Unknown Type = iota
	File
	Score
	ShortCut
	ShortCutCell
	ScoreCol
	ScorePart
	ScoreColPart
	FileCtx
	ScoreCtx
	ScoreColCtx
	ShortCutCtx
)

func (t Type) String() string {
	switch t {
	case File:
		return "File"
	case Score:
		return "Score"
	case ShortCut:
		return "ShortCut"
	case ShortCutCell:
		return "ShortCutCell"
	case ScoreCol:
		return "ScoreCol"
	case ScorePart:
		return "ScorePart"
	case ScoreColPart:
		return "ScoreColPart"
	case FileCtx:
		return "FileCtx"
	case ScoreCtx:
		return "ScoreCtx"
	case ScoreColCtx:
		return "ScoreColCtx"
	case ShortCutCtx:
		return "ShortCutCtx"
	default:
		return "Unknown"
	}
}

func IsReference(str string) bool {
	_, err := parse(str)
	return err == nil
}

func NewFileCtx(file string) Reference {
	return Reference{
		File: file,
		Type: FileCtx,
	}
}

func NewShortCutCtx(file, table string) Reference {
	return Reference{
		File:  file,
		Table: table,
		Type:  ShortCutCtx,
	}
}

func NewScoreColCtx(file, table, col string) Reference {
	return Reference{
		File:  file,
		Table: table,
		Type:  ScoreColCtx,
		Col:   col,
	}
}

func NewScoreCtx(file, table string) Reference {
	return Reference{
		File:  file,
		Table: table,
		Type:  ScoreCtx,
	}
}

// ParseContext tries to guess the context from the given string.
// Returns an error, if the context is not complete.
func ParseContext(ctx string) (r *Reference, err error) {
	r, err = parse(ctx)
	if err != nil {
		return
	}

	if r.File == "" {
		err = fmt.Errorf("missing file name")
		return
	}

	defer func() {
		if err != nil {
			if !r.Valid() {
				err = fmt.Errorf("%s is incomplete", r.Type.String())
			}
		}
	}()

	r.Row = ""
	r.Part = ""

	if r.Table == "" {
		r.Type = FileCtx
		r.Col = ""
		return
	}

	switch r.Table[0] {
	case '.':
		r.Type = ShortCutCtx
		r.Col = ""
	case '=':
		if r.Col == "" {
			r.Type = ScoreCtx
		} else {
			r.Type = ScoreColCtx
		}
	default:
		err = ErrInvalidRefString
	}
	return

}

func Parse(ref string) (r *Reference, err error) {
	r, err = parse(ref)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var ErrInvalidRefString = fmt.Errorf("invalid refence string")

/*
'piano=SCORE.timbre[CHORUS]
'piano=SCORE.timbre
'piano=SCORE
'piano
=SCORE.timbre[CHORUS]
=SCORE.timbre
=SCORE
=.timbre[CHORUS]
=.timbre
'piano=.timbre[CHORUS]
'piano=.timbre
'piano=SCORE[CHORUS]

  oder

'UVI.ks.tremolo.cello
'UVI.ks.tremolo
.ks.tremolo.cello
.ks.tremolo
*/

func parse(ref string) (*Reference, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty string is not a reference")
	}

	if ref[0] == '\'' {

		idx := strings.LastIndex(ref, "/")

		if idx > 2 {
			if len(ref) > idx {
				r, err := parse("'" + ref[idx+1:])
				if err != nil {
					return nil, err
				}
				r.File = filepath.Join(ref[1:idx], r.File)
				return r, nil
			} else {
				return nil, fmt.Errorf("invalid reference %q", ref)
			}
		}
	}

	var state int

	var start = 0
	var inFile = 1
	var inScore = 2
	var inShortCut = 3
	var inRow = 4
	var inCol = 5
	var inPart = 6
	var end = 7
	var r Reference
	var hasError bool

	for _, ch := range ref {
		if hasError {
			break
		}

		switch state {
		case start:
			switch ch {
			case '\'':
				state = inFile
			case '=':
				state = inScore
				r.Table = "="
			case '.':
				state = inShortCut
				r.Table = "."
			case '[':
				state = inPart
			case ']':
				hasError = true
			default:
				hasError = true
			}
		case inFile:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				state = inScore
				r.Table = "="
			case '.':
				state = inShortCut
				r.Table = "."
			case '[':
				state = inPart
			case ']':
				hasError = true
			default:
				r.File = r.File + string(ch)
			}
		case inScore:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				hasError = true
			case '.':
				state = inCol
				r.Col = "."
			case '[':
				state = inPart
			case ']':
				hasError = true
			default:
				r.Table = r.Table + string(ch)
			}
		case inShortCut:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				hasError = true
			case '.':
				r.Row = "."
				state = inRow
			case '[':
				hasError = true
			case ']':
				hasError = true
			default:
				r.Table = r.Table + string(ch)
			}
		case inRow:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				hasError = true
			case '.':
				r.Col = "."
				state = inCol
			case '[':
				hasError = true
			case ']':
				hasError = true
			default:
				r.Row = r.Row + string(ch)
			}
		case inCol:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				hasError = true
			case '.':
				r.Col = r.Col + string(ch)
			case '[':
				state = inPart
			case ']':
				hasError = true
			default:
				r.Col = r.Col + string(ch)
			}
		case inPart:
			switch ch {
			case '\'':
				hasError = true
			case '=':
				hasError = true
			case '.':
				hasError = true
			case '[':
				hasError = true
			case ']':
				state = end
			default:
				r.Part = r.Part + string(ch)
			}
		default:
			hasError = true
		}

	}

	if state == inPart || hasError {
		return nil, ErrInvalidRefString
	}

	return &r, nil

}

func (r *Reference) parseFileCtx(ctx Reference) error {
	/* TODO
	inside the file context, there could be only
	whole files imported (Type File)
	or a single score of a file (Type Score)
	or a single shortcut table of a file (Type ShortCut)
	*/

	// in each case, we need a different file name
	if r.File == "" || r.File == ctx.File {
		return fmt.Errorf("can't import from same file in file context")
	}

	if r.Table == "" {
		r.Type = File
	} else {
		switch r.Table[0] {
		case '=':
			r.Type = Score
		case '.':
			r.Type = ShortCut
		default:
			return fmt.Errorf("invalid table name: %#v", r.Table)
		}
	}

	return nil
}

func (r *Reference) parseScoreCtx(ctx Reference) error {
	/* TODO
	inside the score context, there could be only
	a single score of a file (Type Score)
	or a part of a single score of a file (Type ScorePart)
	*/

	if r.Part != "" {
		r.Type = ScorePart
	} else {
		if r.Table == "=" || r.Table == ctx.Table {
			if r.File == "" || r.File == ctx.File {
				return fmt.Errorf("can't import from same score and same file in score context")
			}
		}
		r.Type = Score
	}

	return nil
}

func (r *Reference) parseScoreColCtx(ctx Reference) error {

	/* TODO
	inside the score col context, there could be only
	a score column of a file (Type ScoreCol)
	the part of a score column (Type ScoreColPart)
	and a shortcut entry  (Type ShortCutCell)
	*/

	if r.Table == "" && r.Part == "" && r.File == "" {
		return fmt.Errorf("invalid reference in column context: must have table or '='")
	}

	if r.File == "" || r.File == ctx.File {
		if r.Part == "" && (r.Table == "=" || r.Table == ctx.Table) && (r.Col == "." || r.Col == ctx.Col) {
			return fmt.Errorf("invalid reference in column context: can't refer to same col in same score in same file")
		}
	}

	if r.Table == "" {
		if r.Part != "" {
			r.Type = ScoreColPart
		} else {
			r.Type = ScoreCol
		}
		return nil
	}

	switch r.Table[0] {
	case '.':
		if r.Row == "" {
			return fmt.Errorf("need row in shortcut reference")
		}
		r.Type = ShortCutCell
	case '=':
		if r.Part != "" {
			r.Type = ScoreColPart
		} else {
			r.Type = ScoreCol
		}
	default:
		return fmt.Errorf("invalid table name")
	}

	return nil
}

func (r *Reference) parseShortCutCtx() error {
	/* TODO
	inside the shortcut context, there could be only
	a single shortcut entry (Type ShortCutCell)
	*/
	if r.Row == "" {
		return fmt.Errorf("need row in shortcut reference")
	}
	r.Type = ShortCutCell
	return nil
}

type Reference struct {
	Type
	File  string
	Table string
	Row   string // empty for score tables
	Col   string
	Part  string // empty for shortcut tables
}

func (r Reference) Dup() Reference {
	return r
}

// Valid checks, if the reference is valid
func (r Reference) Valid() bool {
	switch r.Type {
	case Unknown:
		return false
	case File:
		return r.IsComplete()
	case Score:
		if !r.IsComplete() {
			return false
		}
		if r.Table[0] != '=' {
			return false
		}
	case ShortCut:
		if !r.IsComplete() {
			return false
		}
		if r.Table[0] != '.' {
			return false
		}
	case ShortCutCell:
		if r.Table == "" || r.Table[0] != '.' {
			return false
		}
	case ScoreCol:
		if r.Table == "" || r.Table[0] != '=' {
			return false
		}
		if r.File == "" && r.Table == "=" && r.Col == "" {
			return false
		}
	case ScorePart:
		if r.Table == "" || r.Table[0] != '=' {
			return false
		}
		if r.Part == "" {
			return false
		}
	case ScoreColPart:
		if r.Table == "" || r.Table[0] != '=' {
			return false
		}
		if r.Part == "" {
			return false
		}
	case FileCtx:
		return r.IsComplete()
	case ScoreCtx:
		return r.IsComplete()
	case ScoreColCtx:
		return r.IsComplete()
	case ShortCutCtx:
		return r.IsComplete()
	default:
		return false
	}
	return true
}

// Complete tries to complete the reference via the given context reference
func (r *Reference) Complete(ctx Reference) error {
	switch ctx.Type {
	case FileCtx:
		err := r.parseFileCtx(ctx)
		if err != nil {
			return err
		}
		switch r.Type {
		case File:
			if r.File == "" {
				return fmt.Errorf("can't import file within itself")
			}
			return nil
		case Score:
			if r.File == "" {
				return fmt.Errorf("can't import file within itself")
			}
			if r.Table == "=" {
				return fmt.Errorf("can't complete reference table name in file context")
			}
			return nil
		case ShortCut:
			if r.File == "" {
				return fmt.Errorf("can't import file within itself")
			}
			if r.Table == "" {
				return fmt.Errorf("can't complete reference table name in file context")
			}
			return nil
		default:
			return fmt.Errorf("invalid reference type: %s", ctx.Type)
		}
	case ScoreCtx:
		err := r.parseScoreCtx(ctx)
		if err != nil {
			return err
		}
		//fmt.Printf("ctx: %s, r.Type: %s %#v\n", ctx.Type.String(), r.Type.String(), r)
		switch r.Type {
		case Score, ScorePart:
			if r.File == "" && r.Table == "=" && r.Part == "" {
				return fmt.Errorf("can't complete reference table name and file name in score context")
			}
			if r.File == "" {
				r.File = ctx.File
				if r.Table == "=" || r.Table == "" {
					r.Table = ctx.Table
				}
			} else {
				if r.Table == "=" || r.Table == "" {
					r.Table = "=SCORE"
				}
			}

			return nil
		default:
			return fmt.Errorf("invalid reference type: %s", ctx.Type)
		}

	case ScoreColCtx:
		err := r.parseScoreColCtx(ctx)
		if err != nil {
			return err
		}
		//fmt.Printf("completing type %s within %#v\n", r.Type.String(), r)
		switch r.Type {
		case ShortCutCell:
			if r.File == "" {
				r.File = ctx.File
			}

			if r.Col == "." {
				r.Col = ctx.Col
			}

			return nil
		case ScoreCol, ScoreColPart:
			if r.File == "" {
				r.File = ctx.File
				if r.Table == "" || r.Table == "=" {
					r.Table = ctx.Table
					if r.Col == "" && r.Part != "" {
						r.Col = ctx.Col
					}
				}

			}

			if r.Table == "" || r.Table == "=" {
				r.Table = "=SCORE"
			}

			if r.Col == "." {
				r.Col = ctx.Col
			}

			if r.Type == ScoreCol && r.File == ctx.Table && r.Table == ctx.Table && r.Col == ctx.Col {
				return fmt.Errorf("can't complete reference table name and file name and col name in score col context")
			}

			return nil
		default:
			return fmt.Errorf("invalid reference type: %s", ctx.Type)
		}
	case ShortCutCtx:
		err := r.parseShortCutCtx()
		if err != nil {
			return err
		}
		switch r.Type {
		case ShortCutCell:
			if r.File == "" {
				r.File = ctx.File
			}
			/*
				if r.Col == "" {
					r.Col = ctx.Col
				}
			*/
			return nil
		default:
			return fmt.Errorf("invalid reference type: %s", ctx.Type)
		}
	default:
		return fmt.Errorf("invalid context type: %s", ctx.Type)

	}
}

// IsComplete checks, if the reference is complete
func (r Reference) IsComplete() bool {
	switch r.Type {
	case Unknown:
		return false
	case File:
		if r.File == "" {
			return false
		}
	case Score:
		if r.File == "" || r.Table == "" {
			return false
		}
	case ShortCut:
		if r.File == "" || r.Table == "" {
			return false
		}
	case ShortCutCell:
		if r.File == "" || r.Table == "" || r.Row == "" {
			return false
		}
	case ScoreCol:
		if r.File == "" || r.Table == "" {
			return false
		}
	case ScorePart:
		if r.File == "" || r.Table == "" || r.Part == "" {
			return false
		}
	case ScoreColPart:
		if r.File == "" || r.Table == "" || r.Part == "" {
			return false
		}
	case FileCtx:
		if r.File == "" {
			return false
		}
	case ScoreCtx:
		if r.File == "" {
			return false
		}
		if r.Table == "" {
			return false
		}
	case ScoreColCtx:
		if r.File == "" {
			return false
		}
		if r.Table == "" {
			return false
		}

	case ShortCutCtx:
		if r.File == "" {
			return false
		}
		if r.Table == "" {
			return false
		}
		if r.Row == "" {
			return false
		}
		if r.Col == "" {
			return false
		}
	default:
		return false
	}
	return true
}

// Cols returns the columns
func (r Reference) Cols() []string {
	if r.Col == "" {
		return nil
	}
	return strings.Split(r.Col, " ")
}

/*
'piano=SCORE.timbre[CHORUS]
'piano=SCORE.timbre
'piano=SCORE
'piano
=SCORE.timbre[CHORUS]
=SCORE.timbre
=SCORE
=.timbre[CHORUS]
=.timbre
'piano=.timbre[CHORUS]
'piano=.timbre
'piano=SCORE[CHORUS]

  oder

'UVI.ks.tremolo.cello
'UVI.ks.tremolo
.ks.tremolo.cello
.ks.tremolo
*/
func (r Reference) String() string {
	var bf bytes.Buffer

	if r.File != "" {
		bf.WriteString("'" + strings.ReplaceAll(r.File, "\\", "/"))
	}

	if r.Table != "" {
		bf.WriteString(r.Table)
		if r.Table[0] == '.' {
			bf.WriteString(r.Row)
		}
	}

	bf.WriteString(r.Col)

	if r.Part != "" {
		bf.WriteString("[" + r.Part + "]")
	}

	return bf.String()
}
