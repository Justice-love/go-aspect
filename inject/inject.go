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
			if iline == jline && ai.Point.mode == aj.Point.mode {
				panic("duplicate")
			}
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
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	line := 0
	if aspect.Function.ReturnLine > 0 {
		line = aspect.Function.ReturnLine - 1
	} else {
		line = aspect.Function.FuncEndLine
	}
	util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code+"\n", aspect), line)
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
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code+"\n", aspect), aspect.Function.FuncLine)
}

func (b BeforeInjectFile) FunctionLine(aspect *Aspect) int {
	return aspect.Function.FuncLine
}

func (b BeforeInjectFile) Name() string {
	return "Before"
}

func (d DeferInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	code := `
	defer func() {
		` + aspect.Point.code + `
	}()` + "\n"

	util.InsertStringToFile(sourceStruct.Path, bindParam(code, aspect), aspect.Function.FuncLine)
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
	for _, one := range aspect.Point.imports {
		if one.SourceContain && one.SourceTag != one.ImportTag {
			e := one.ImportEndTerm
			if one.ImportTag != "" {
				e = one.ImportTag
			}
			n := one.SourceTag
			if one.SourceTag == "" {
				n = one.ImportEndTerm
			}
			code = strings.ReplaceAll(code, e+".", n+".")
		}
	}
	return code
}

type AroundInjectFile struct {
}

func (a AroundInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	name := util.ReplaceFunctionName(sourceStruct.Path, aspect.Function.FuncName, aspect.Function.NameLine)
	code := aspect.Function.FuncString + "\n"
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTarget(aspect.Function, name) + "\n"
	} else {
		code += "\t" + aroundTarget(aspect.Function, name) + "\n"
	}
	code += "}\n"
	util.Append(sourceStruct.Path, code)
}

func aroundTarget(function *parse.FuncStruct, name string) string {
	if function.Receiver != nil {
		return function.Receiver.Alias + "." + name + targetParam(function)
	} else {
		return name + targetParam(function)
	}
}

func targetParam(function *parse.FuncStruct) string {
	code := "("
	for i, one := range function.Params {
		if i == len(function.Params)-1 {
			code += one.Name
		} else {
			code += one.Name + ","
		}
	}
	code += ")"
	return code
}

func (a AroundInjectFile) FunctionLine(*Aspect) int {
	return int(^uint(0) >> 1)
}

func (a AroundInjectFile) Name() string {
	return "Around"
}
