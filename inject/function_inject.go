package inject

import (
	"github.com/Justice-love/go-aspect/parse"
	"log"
	"sort"
)

type FunctionInject struct {
	f       *parse.FuncStruct
	aspects []*Aspect
}

func funcInjects(advice *Advice) {
	fm := make(map[*parse.FuncStruct]*FunctionInject)
	for _, aspect := range advice.Aspect {
		f, ok := fm[aspect.Function]
		if !ok {
			f = &FunctionInject{f: aspect.Function}
			fm[aspect.Function] = f
		}
		f.aspects = append(f.aspects, aspect)
	}
	for _, v := range fm {
		sort.SliceStable(v.aspects, func(i, j int) bool {
			if v.aspects[i].Function == v.aspects[j].Function {
				log.Fatal("duplicate func inject")
			}
			return v.aspects[i].Point.mode.Sort() < v.aspects[j].Point.mode.Sort()
		})
		advice.fi = append(advice.fi, v)
	}
	sort.SliceStable(advice.fi, func(i, j int) bool {
		return advice.fi[i].f.FuncLine > advice.fi[j].f.FuncLine
	})
}
