package inject

import (
	"fmt"
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
	for k, v := range adviceMap {
		sort.Slice(v.Aspect, func(i, j int) bool {
			ai := v.Aspect[i]
			aj := v.Aspect[j]
			if ai.Function == aj.Function {
				log.Fatal("duplicate function aspect")
			}
			return ai.Function.FuncLine > aj.Function.FuncLine
		})
		util.InitFile(k.XgcPath, k.PackageStr)
		for _, one := range v.Aspect {
			injectFun(k, one)
		}
		injectImports(v)
	}
}

func injectImports(advice *Advice) {
	importMap := make(map[string]*parse.ImportStruct)
	for _, one := range advice.Aspect {
		imports := filterImports(one.Point)
		for _, i := range imports {
			importMap[i.ImportString] = i
		}
	}
	str := ""
	for _, one := range importMap {
		str += fmt.Sprint("import\t", one.ImportTag, " ", "\"", one.ImportString, "\"\n")
	}
	str += "\n"
	util.InsertStringToFile(advice.Source.XgcPath, str, 2)
}

func injectFun(source *parse.SourceStruct, aspect *Aspect) {
	log.Debugf("inject:%s %s", aspect.Point.mode.Name(), aspect.Function.FuncName)
	aspect.Point.mode.InjectFunc(source, aspect)
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

type CodeInjectInterface interface {
	InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect)
	Name() string
	Sort() int
}

var injectMap = map[string]CodeInjectInterface{
	"before": &BeforeInjectFile{},
	"after":  &AfterInjectFile{},
	//"defer":  &DeferInjectFile{},
	"around": &AroundInjectFile{},
}

type DeferInjectFile struct {
}

type AfterInjectFile struct {
}

type BeforeInjectFile struct {
}

func (a *AfterInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	after := bindParam(aspect.Point.code+"\n", aspect)
	in := `
	defer func() {
		` + after + `
	}()` + "\n"
	_ = util.ReplaceFunctionName(sourceStruct.Path, aspect.Function.FuncName, aspect.Function.NameLine)
	code := aspect.Function.FuncString + "\n"
	code += in
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTarget(aspect.Function, util.Prefix+aspect.Function.FuncName) + "\n"
	} else {
		code += "\t" + aroundTarget(aspect.Function, util.Prefix+aspect.Function.FuncName) + "\n"
	}
	code += "}"
	util.Append(sourceStruct.XgcPath, code)
}

func (a *AfterInjectFile) Name() string {
	return "After"
}

func (b *BeforeInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	before := bindParam(aspect.Point.code+"\n", aspect)
	_ = util.ReplaceFunctionName(sourceStruct.Path, aspect.Function.FuncName, aspect.Function.NameLine)
	code := aspect.Function.FuncString + "\n"
	code += before
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTarget(aspect.Function, util.Prefix+aspect.Function.FuncName) + "\n"
	} else {
		code += "\t" + aroundTarget(aspect.Function, util.Prefix+aspect.Function.FuncName) + "\n"
	}
	code += "}"
	util.Append(sourceStruct.XgcPath, code)
}

func (b *BeforeInjectFile) Name() string {
	return "Before"
}

func (d *DeferInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	code := `
	defer func() {
		` + aspect.Point.code + `
	}()` + "\n"

	util.InsertStringToFile(sourceStruct.Path, bindParam(code, aspect), aspect.Function.FuncLine)
}

func (d *DeferInjectFile) Name() string {
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

type AroundInjectFile struct {
}

func (a *AroundInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}
	invoke := aroundTarget(aspect.Function, util.Prefix+aspect.Function.FuncName) + "\n"
	around := strings.Replace(aspect.Point.code, "invoke()", invoke, 1)
	around = bindParam(around, aspect)
	_ = util.ReplaceFunctionName(sourceStruct.Path, aspect.Function.FuncName, aspect.Function.NameLine)
	code := aspect.Function.FuncString + "\n"
	code += around
	code += "\n}"
	util.Append(sourceStruct.XgcPath, code)
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

func (a *AroundInjectFile) Name() string {
	return "Around"
}

func (d *DeferInjectFile) Sort() int {
	return 1
}

func (a *AfterInjectFile) Sort() int {
	return 4
}

func (b *BeforeInjectFile) Sort() int {
	return 2
}

func (a *AroundInjectFile) Sort() int {
	return 3
}
