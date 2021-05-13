package inject

import (
	"eddy.org/go-aspect/parse"
	"eddy.org/go-aspect/util"
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
		if len(v.Aspect[0].Point.imports) > 0 {
			v.Source.InjectImport(v.Source, v.Aspect[0].Point.imports)
		}

	}
}

type InjectInterface interface {
	InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect)
	FunctionLine(aspect *Aspect) int
	Name() string
}

var injectMap = map[string]InjectInterface{
	"before": BeforeInjectFile{},
	"after":  AfterInjectFile{},
}

type AfterInjectFile struct {
}

type BeforeInjectFile struct {
}

func (a AfterInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code+"\n", aspect), aspect.Function.FuncEndLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (a AfterInjectFile) FunctionLine(aspect *Aspect) int {
	return aspect.Function.FuncEndLine
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
