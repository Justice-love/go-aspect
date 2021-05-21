package inject

import (
	"github.com/Justice-love/go-aspect/parse"
	"github.com/Justice-love/go-aspect/util"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
)

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
			iline := ai.Point.mode.FunctionLine(ai)
			jline := aj.Point.mode.FunctionLine(aj)
			return iline > jline
		})
		for _, one := range v.Aspect {
			log.Debugf("inject:%s %s", one.Point.mode.Name(), one.Function.FuncName)
			one.Point.mode.InjectFunc(v.Source, one)
		}
		for _, one := range v.Aspect {
			imports := filterImports(one.Point)
			if len(imports) == 0 {
				continue
			} else {
				v.Source.InjectImport(v.Source, imports)
			}
		}

	}
}

func filterImports(point *Point) (imports []*parse.ImportStruct) {
	for _, one := range point.imports {
		if one.ImportTag != "" && strings.Contains(point.code, one.ImportTag+".") {
			imports = append(imports, one)
		} else if strings.Contains(point.code, one.ImportEndTerm+".") {
			imports = append(imports, one)
		}
	}
	return
}

type InjectInterface interface {
	InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect)
	FunctionLine(aspect *Aspect) int
	Name() string
}

var injectMap = map[string]InjectInterface{
	"before": BeforeInjectFile{},
	"after":  AfterInjectFile{},
	"defer":  DeferInjectFile{},
}

type DeferInjectFile struct {
}

type AfterInjectFile struct {
}

type BeforeInjectFile struct {
}

func (a AfterInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	line := 0
	if aspect.Function.ReturnLine > 0 {
		line = aspect.Function.ReturnLine - 1
	} else {
		line = aspect.Function.FuncEndLine
	}
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code+"\n", aspect), line)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (a AfterInjectFile) FunctionLine(aspect *Aspect) int {
	if aspect.Function.ReturnLine > 0 {
		return aspect.Function.ReturnLine - 1
	} else {
		return aspect.Function.FuncEndLine
	}
}

func (a AfterInjectFile) Name() string {
	return "After"
}

func (b BeforeInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code+"\n", aspect), aspect.Function.FuncLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (b BeforeInjectFile) FunctionLine(aspect *Aspect) int {
	return aspect.Function.FuncLine
}

func (b BeforeInjectFile) Name() string {
	return "Before"
}

func (d DeferInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	code := `
	defer func() {
		` + aspect.Point.code + `
	}()` + "\n"

	err := util.InsertStringToFile(sourceStruct.Path, bindParam(code, aspect), aspect.Function.FuncLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (d DeferInjectFile) FunctionLine(aspect *Aspect) int {
	return aspect.Function.FuncLine
}

func (d DeferInjectFile) Name() string {
	return "Defer"
}

func bindParam(code string, aspect *Aspect) string {
	if aspect.Point.injectReceiver != nil && aspect.Function.Receiver.Alias != "" {
		code = strings.ReplaceAll(code, "{{"+aspect.Point.injectReceiver.Receiver+"}}", aspect.Function.Receiver.Alias)
	}
	for i, p := range aspect.Function.Params {
		pointParam := aspect.Point.params[i]
		if p.Name == pointParam.Name {
			continue
		}
		code = strings.ReplaceAll(code, "{{"+pointParam.Name+"}}", p.Name)
	}
	return code
}
