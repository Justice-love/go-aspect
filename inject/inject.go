package inject

import (
	"bytes"
	"fmt"
	"github.com/Justice-love/go-aspect/parse"
	"github.com/Justice-love/go-aspect/writer"
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
				log.Panic("duplicate function aspect")
			}
			return ai.Function.FuncLine > aj.Function.FuncLine
		})
		writer.InitFile(k.XgcPath, k.PackageStr)
		injectImports(v)
		for _, one := range v.Aspect {
			injectFun(k, one)
		}
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
	writer.Append(advice.Source.XgcPath, str)
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
}

var injectMap = map[string]CodeInjectInterface{
	"before": &BeforeInjectFile{},
	"after":  &AfterInjectFile{},
	"around": &AroundInjectFile{},
	"test":   &TestInjectFile{},
}

type AfterInjectFile struct {
}

type BeforeInjectFile struct {
}

func (a *AfterInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}

	codeStruct := SourceStructStr(aspect.Function)
	writer.Append(sourceStruct.XgcPath, codeStruct)

	after := bindParam(aspect.Point.code+"\n", aspect)
	in := `
	defer func() {
		` + after + `
	}()` + "\n"
	if aspect.Function.Receiver != nil {
		writer.ReplaceReceiver(sourceStruct.Path, aspect.Function.Receiver.Receiver, parse.FunctionReceiverLineNum(aspect.Function.LineString))
	} else {
		writer.AddReceiver(sourceStruct.Path, aspect.Function.FuncName, parse.FunctionNameLineNum(aspect.Function.LineString))
	}
	code := SourceFunctionStr(aspect.Function) + "\n"
	code += in
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	} else {
		code += "\t" + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	}
	code += "}"
	writer.Append(sourceStruct.XgcPath, code)
}

func (a *AfterInjectFile) Name() string {
	return "After"
}

func (b *BeforeInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}

	codeStruct := SourceStructStr(aspect.Function)
	writer.Append(sourceStruct.XgcPath, codeStruct)

	before := bindParam(aspect.Point.code+"\n", aspect)
	if aspect.Function.Receiver != nil {
		writer.ReplaceReceiver(sourceStruct.Path, aspect.Function.Receiver.Receiver, parse.FunctionReceiverLineNum(aspect.Function.LineString))
	} else {
		writer.AddReceiver(sourceStruct.Path, aspect.Function.FuncName, parse.FunctionNameLineNum(aspect.Function.LineString))
	}
	code := SourceFunctionStr(aspect.Function) + "\n"
	code += before
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	} else {
		code += "\t" + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	}
	code += "}"
	writer.Append(sourceStruct.XgcPath, code)
}

func (b *BeforeInjectFile) Name() string {
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

type AroundInjectFile struct {
}

func (a *AroundInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}

	codeStruct := SourceStructStr(aspect.Function)
	writer.Append(sourceStruct.XgcPath, codeStruct)

	invoke := aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	around := strings.Replace(aspect.Point.code, "invoke()", invoke, 1)
	around = bindParam(around, aspect)
	if aspect.Function.Receiver != nil {
		writer.ReplaceReceiver(sourceStruct.Path, aspect.Function.Receiver.Receiver, parse.FunctionReceiverLineNum(aspect.Function.LineString))
	} else {
		writer.AddReceiver(sourceStruct.Path, aspect.Function.FuncName, parse.FunctionNameLineNum(aspect.Function.LineString))
	}
	code := SourceFunctionStr(aspect.Function) + "\n"
	code += around
	code += "\n}"
	writer.Append(sourceStruct.XgcPath, code)
}

func aroundTarget(function *parse.FuncStruct, name string) string {
	if function.Receiver != nil && function.Receiver.Alias == "" {
		return "x." + name + targetParam(function)
	} else if function.Receiver != nil {
		return function.Receiver.Alias + "." + name + targetParam(function)
	} else {
		return name + targetParam(function)
	}
}

func aroundTargetWithNewReceiver(function *parse.FuncStruct, name string) string {
	return ReceiverNew(function) + name + targetParam(function)
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

func SourceFunctionStr(f *parse.FuncStruct) string {
	if (f.Receiver != nil && f.Receiver.Alias != "") || f.Receiver == nil {
		return f.FuncString
	}
	r := f.Receiver
	if r.Pointer {
		return strings.Replace(f.FuncString, "*"+r.Receiver, "x *"+r.Receiver, 1)
	} else {
		return strings.Replace(f.FuncString, r.Receiver, "x "+r.Receiver, 1)
	}
}

func SourceStructStr(f *parse.FuncStruct) string {
	var buffer bytes.Buffer
	typeName := writer.Prefix + f.FuncName
	if f.Receiver != nil {
		typeName = writer.Prefix + f.Receiver.Receiver
	}
	buffer.WriteString("type ")
	buffer.WriteString(typeName)
	buffer.WriteString(" struct{\n")
	if f.Receiver != nil {
		buffer.WriteString("\t")
		if f.Receiver.Pointer {
			buffer.WriteString("*")
		}
		buffer.WriteString(f.Receiver.Receiver)
		buffer.WriteString("\n")
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

func ReceiverNew(f *parse.FuncStruct) string {
	if f.Receiver != nil && f.Receiver.Alias != "" {
		return fmt.Sprint("(&", writer.Prefix, f.Receiver.Receiver, "{", f.Receiver.Alias, "}).")
	} else if f.Receiver != nil {
		return fmt.Sprint("(&", writer.Prefix, f.Receiver.Receiver, "{", "x", "}).")
	} else {
		return fmt.Sprint("(&", writer.Prefix, f.FuncName, "{", "}).")
	}
}
