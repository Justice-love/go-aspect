package inject

import (
	"eddy.org/go-aspect/parse"
	"eddy.org/go-aspect/util"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
)

func BeforeInjectCode(advices []*Advice) {
	adviceMap := make(map[*parse.SourceStruct]*Advice)
	for _, one := range advices {
		if advice, ok := adviceMap[one.Source]; ok {
			advice.Aspect = append(advice.Aspect, one.Aspect...)
		} else {
			adviceMap[one.Source] = one
		}
	}
	for _, v := range adviceMap {
		sort.SliceStable(v.Aspect, func(i, j int) bool {
			ai := v.Aspect[i]
			aj := v.Aspect[j]
			return ai.Function.FuncLine > aj.Function.FuncLine
		})
		for _, one := range v.Aspect {
			ctxName := contextName(one.Function.Params)
			err := util.InsertStringToFile(v.Source.Path, bindParam(one.Point.code, ctxName)+"\n", one.Function.FuncLine)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
		if len(v.Aspect[0].Point.imports) > 0 {
			v.Source.InjectImport(v.Source, v.Aspect[0].Point.imports)
		}

	}
}

func DoInjectCode(advices []*Advice) {
	adviceMap := make(map[*parse.SourceStruct]*Advice)
	for _, one := range advices {
		if advice, ok := adviceMap[one.Source]; ok {
			advice.Aspect = append(advice.Aspect, one.Aspect...)
		} else {
			adviceMap[one.Source] = one
		}
	}
	for _, v := range adviceMap {
		sort.SliceStable(v.Aspect, func(i, j int) bool {
			ai := v.Aspect[i]
			aj := v.Aspect[j]
			return ai.Function.FuncLine > aj.Function.FuncLine
		})
		for _, one := range v.Aspect {
			one.Point.mode(v.Source, one)
		}
		if len(v.Aspect[0].Point.imports) > 0 {
			v.Source.InjectImport(v.Source, v.Aspect[0].Point.imports)
		}

	}
}

func bindParam(code string, name string) string {
	if name == "ctx" {
		return code
	} else {
		return strings.ReplaceAll(code, "ctx", name)
	}
}

func contextName(params []*parse.ParamStruct) string {
	for _, one := range params {
		if one.Context {
			return one.Name
		}
	}
	log.Fatal("no context param")
	return ""
}
