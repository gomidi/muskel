package items

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ItemGroup []*Item

/*
- slicing mit []
- dynamik change mit + und -
- diatonische transposition mit ^
- chromatische transposition mit # und b
- octavierungen mit " und '
- microtimingverschiebung mit > und <
- zufallswahrscheinlichkeit mit xx%
- zufällige auswahlmit %()
*/

/*
items.Token       -> ItemGroup
items.MultiItem   -> ItemGroup
items.Scale       -> ItemGroup

items.Pattern     -> EventSequence
items.PartRepeat  -> EventSequence
items.BarRepeater -> EventSequence
*/

type itemGroupModifier struct {
	Params      []string
	Slice       [2]int
	SingleSlice int
	DynamicAdd  string
	ScaleMove   int8
	PosShift    int
	Exploded    bool
}

func (i *itemGroupModifier) string() string {
	var bd strings.Builder

	if len(i.Params) > 0 {
		bd.WriteString("(" + strings.Join(i.Params, ",") + ")")
	}

	if i.ScaleMove != 0 {
		fmt.Fprintf(&bd, "^%v", i.ScaleMove)
	}

	if i.SingleSlice >= 0 {
		fmt.Fprintf(&bd, "[%v]", i.SingleSlice)
	}

	if i.Slice[0] >= 0 || i.Slice[1] >= 0 {
		bd.WriteString("[")

		if i.Slice[0] >= 0 {
			bd.WriteString(fmt.Sprintf("%v", i.Slice[0]))
		}

		bd.WriteString(":")

		if i.Slice[1] >= 0 {
			bd.WriteString(fmt.Sprintf("%v", i.Slice[1]))
		}

		bd.WriteString("]")
	}

	if i.DynamicAdd != "" {
		bd.WriteString(i.DynamicAdd)
	}

	if i.PosShift != 0 {
		if i.PosShift < 0 {
			for j := i.PosShift; j < 0; j++ {
				bd.WriteRune('<')
			}
		} else {
			for j := i.PosShift; j > 0; j-- {
				bd.WriteRune('>')
			}
		}
	}

	return bd.String()
}

var regItemModStr = "^(" +
	regexp.QuoteMeta("^") +
	"{1,2}[-0-9]+){0,1}([" +
	regexp.QuoteMeta("-+") +
	"]*)([<>]{0,3})"

var regItemMod = regexp.MustCompile(regItemModStr)

func (i *itemGroupModifier) parseExploded(data string) (rest string) {
	if idx := strings.Index(data, "..."); idx >= 0 && idx+3 == len(data) {
		i.Exploded = true
		return data[:idx]
	}
	return data
}

func (i *itemGroupModifier) Reset() {
	i.Slice = [2]int{-1, -1}
	i.SingleSlice = -1
	i.Params = nil
	i.ScaleMove = 0
	i.PosShift = 0
}

func (i *itemGroupModifier) parse(data string) (err error) {
	slice := ""
	params := ""
	i.Reset()

	if idx := strings.Index(data, "("); idx >= 0 {
		lidx := strings.LastIndex(data, ")")
		params = strings.TrimSpace(data[idx+1 : lidx])
		data = data[lidx+1:]
	}

	if params != "" {
		i.Params = splitParams(params)
	}

	//fmt.Printf("parse: %q params: %v\n", data, params)

	if idx := strings.Index(data, "["); idx >= 0 {
		lidx := strings.LastIndex(data, "]")
		slice = strings.TrimSpace(data[idx+1 : lidx])
		data = data[lidx+1:]
	}

	if slice != "" {
		sl := strings.Split(slice, ":")
		if len(sl) == 1 {
			i.SingleSlice, err = strconv.Atoi(slice)
			if err != nil {
				return fmt.Errorf("ERROR in slice of item group: invalid slice %q", "["+slice+"]")
			}
			return nil
		}
		if len(sl) != 2 {
			return fmt.Errorf("ERROR in slice of item group: invalid slice %q", "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			i.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in slice of item group: invalid slice %q", "["+slice+"]")
			}
			i.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in slice of item group: invalid slice %q", "["+slice+"]")
			}
			i.Slice[1] = toI
		}

	}

	mt := regItemMod.FindStringSubmatch(data)

	//fmt.Printf("data: %q mt: %v len(mt): %v\n", data, mt, len(mt))

	if mt[1] != "" {
		dt := mt[1][1:]
		si, err := strconv.Atoi(dt)
		if err != nil {
			return fmt.Errorf("error in scale moving of item group modifier: %q is not a number", dt)
		}

		i.ScaleMove = int8(si)
	}

	i.DynamicAdd = mt[2]
	i.PosShift = 0

	for _, s := range mt[3] {
		switch s {
		case '<':
			i.PosShift -= 1
		case '>':
			i.PosShift += 1
		}
	}
	return nil
}

type ItemGrouperFunc func() ItemGroup

func (e ItemGrouperFunc) Get() ItemGroup {
	return e()
}

type ItemGrouper interface {
	Get() ItemGroup
}

type ItemGroupMiddleWare []func(in ItemGroup) ItemGroup

// first function must retrieve the ItemGroup (receives nil)
func (mw ItemGroupMiddleWare) Get() (es ItemGroup) {
	if len(mw) == 0 {
		return nil
	}

	for _, fn := range mw {
		es = fn(es)
	}

	return
}

func (s ItemGroup) ToParams() Params {
	panic("TODO implement")
}

func (s ItemGroup) Call(params Params) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) Slice(from, to int) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) DiatonicTranspose(steps int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) Transpose(halftones int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) TransposeOctave(octave int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) MircoTimingShift(shift int8) ItemGroup {
	panic("TODO implement")
}

func (s ItemGroup) ChangeDynamic(diff int8) ItemGroup {
	panic("TODO implement")
}
