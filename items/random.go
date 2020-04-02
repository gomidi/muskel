package items

import (
	"fmt"
	"strconv"
	"strings"
)

var _ Item = &RandomProbability{}

type RandomProbability struct {
	Prob             uint8
	Item             Item
	Chosen           bool
	ItemOriginalData string
	//Parser           *Parser
}

func (r RandomProbability) Dup() Item {
	return &r
}

func (r RandomProbability) String() string {
	return fmt.Sprintf("?%v%%%s", r.Prob, r.Item.String())
}

func (rp *RandomProbability) Parse(data string, posIn32th uint) (err error) {
	data = data[1:]
	idx := strings.Index(data, "%")
	if idx < 0 {
		return fmt.Errorf("invalid random value: ?%s", data)
	}

	num := data[:idx]

	var n int
	n, err = strconv.Atoi(num)

	if err != nil {
		return fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100", data)
	}

	rp.Prob = uint8(n)

	if rp.Prob < 0 || rp.Prob > 100 {
		return fmt.Errorf("invalid random value: ?%s, syntax must be ?n%% where n is a number between 0 and 100, but number is %v", data, rp.Prob)
	}

	//rp.Item, err = parseItem(rp.Parser, data[idx+1:], posIn32th)
	rp.Item, err = Parse(data[idx+1:], posIn32th)

	if err != nil {
		return fmt.Errorf("invalid random value item: ?%s", data)
	}
	rp.ItemOriginalData = data[idx+1:]
	return nil
}

type RandomChooser struct {
	Alternatives             []Item
	AlternativesOriginalData []string
	Chosen                   int
	//Parser                   *Parser
}

var _ Item = &RandomChooser{}

func (r RandomChooser) Dup() Item {
	return &r
}

func (r *RandomChooser) Parse(data string, posIn32th uint) (err error) {
	data = data[1:]
	alternatives := strings.Trim(data, "()")
	alt := strings.Split(alternatives, ",")
	for _, a := range alt {
		a = strings.TrimSpace(a)
		var item Item
		if a != "" {
			//item, err = parseItem(r.Parser, a, posIn32th)
			item, err = Parse(a, posIn32th)

			if err != nil {
				return fmt.Errorf("invalid random value item: %s: %v", a, err)
			}
			r.Alternatives = append(r.Alternatives, item)
			r.AlternativesOriginalData = append(r.AlternativesOriginalData, a)
		}
	}
	return nil
}

func (r RandomChooser) String() string {
	var bd strings.Builder

	bd.WriteString("?(")

	for i, it := range r.Alternatives {
		if i > 0 {
			bd.WriteRune(',')
		}

		bd.WriteString(it.String())
	}

	bd.WriteString(")")

	return bd.String()
}
