package parse

import (
	"bufio"
	"eddy.org/go-aspect/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type SourceStruct struct {
	Path         string
	PackageStr   string
	Imports      []*ImportStruct
	Funcs        []*FuncStruct
	InjectImport func(sourceStruct *SourceStruct, imports []*ImportStruct)
	ImportLine   int
}

type ImportStruct struct {
	ImportTag    string
	ImportString string
}

type FuncStruct struct {
	Receiver *ReceiverStruct
	FuncName string
	FuncLine int
	Params   []*ParamStruct
	Context  bool
}

type ReceiverStruct struct {
	Pointer  bool
	Alias    string
	Receiver string
}

type ParamStruct struct {
	Pointer    bool
	Name       string
	ParamAlias string
	ParamType  string
	Context    bool
}

func NewFuncStruct(name string, line int) *FuncStruct {
	return &FuncStruct{
		FuncName: name,
		FuncLine: line,
	}
}

func NewSourceStruct(path string) *SourceStruct {
	return &SourceStruct{Path: path, InjectImport: MultiLineInject}
}

func SourceParse(sourceFile string) *SourceStruct {
	f, err := os.Open(sourceFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func() {
		_ = f.Close()
	}()
	source := NewSourceStruct(sourceFile)
	reader := bufio.NewReader(f)
	line := 0
	for {
		content, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		line++
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "package") {
			source.PackageStr = strings.TrimSpace(strings.ReplaceAll(contentStr, "package", ""))
		} else if strings.HasPrefix(contentStr, "import") {
			imr, inline := importParse(reader, contentStr)
			source.Imports = append(source.Imports, imr...)
			if inline {
				source.ImportLine = line
				source.InjectImport = InlineImportInject
			} else {
				line += len(imr) + 1
				source.ImportLine = line - 1
			}
		} else if strings.HasPrefix(contentStr, "func") {
			source.Funcs = append(source.Funcs, funcParse(reader, contentStr, line))
		}
	}
	if source.Imports == nil {
		source.ImportLine = 2
		source.InjectImport = InlineImportInject
	}
	return source
}

func funcParse(reader *bufio.Reader, str string, line int) *FuncStruct {
	if strings.HasSuffix(strings.TrimSpace(str), "{") {
		return funcInline(str, line)
	} else {
		return funcMultiLine(reader, str)
	}
}

func funcMultiLine(reader *bufio.Reader, str string) *FuncStruct {
	log.Fatal("unsupported")
	return nil
}

func funcInline(str string, line int) (fun *FuncStruct) {
	str = strings.TrimSpace(strings.TrimLeft(str, "func"))
	var (
		receiver *ReceiverStruct
	)
	if strings.HasPrefix(str, "(") {
		receiver = inlineReceiver(str[1:strings.Index(str, ")")])
		str = strings.TrimSpace(str[strings.Index(str, ")")+1:])
	}
	funName := str[:strings.Index(str, "(")]
	paramStr := str[strings.Index(str, "(")+1 : strings.Index(str, ")")]
	params, ctx := inlineParam(paramStr)
	fun = NewFuncStruct(funName, line)
	fun.Receiver = receiver
	fun.Params = params
	fun.Context = ctx
	return
}

func inlineReceiver(s string) *ReceiverStruct {
	var r *ReceiverStruct
	arr := strings.Split(strings.TrimSpace(s), " ")
	if len(arr) == 1 {
		t, p := checkPointer(strings.TrimSpace(arr[0]))
		r = &ReceiverStruct{
			Pointer:  p,
			Alias:    "",
			Receiver: t,
		}
	} else {
		t, p := checkPointer(strings.TrimSpace(arr[1]))
		r = &ReceiverStruct{
			Pointer:  p,
			Alias:    strings.TrimSpace(arr[0]),
			Receiver: t,
		}
	}
	return r
}

func checkPointer(s string) (t string, p bool) {
	if strings.HasPrefix(s, "*") {
		t = s[1:]
		p = true
	} else {
		t = s
	}
	return
}

func inlineParam(str string) (params []*ParamStruct, ctx bool) {
	ps := strings.Split(strings.TrimSpace(str), ",")
	var typeHoldOn []*ParamStruct
	for _, one := range ps {
		kvs := strings.Split(strings.TrimSpace(one), " ")
		if len(kvs) == 1 {
			typeHoldOn = append(typeHoldOn, &ParamStruct{Name: kvs[0]})
		} else {
			types := strings.Split(kvs[1], ".")
			var param *ParamStruct
			if len(types) == 1 {
				param = &ParamStruct{Name: kvs[0], ParamType: types[0]}
			} else {
				t, p := checkPointer(types[1])
				param = &ParamStruct{Name: kvs[0], ParamAlias: types[0], ParamType: t}
				param.Pointer = p
				if types[0] == "context" && types[1] == "Context" {
					ctx = true
					param.Context = true
				}
			}
			params = append(params, param)
			if len(typeHoldOn) > 0 {
				for _, h := range typeHoldOn {
					h.ParamType = param.ParamType
					h.ParamAlias = param.ParamAlias
					h.Context = param.Context
					params = append(params, h)
				}
				typeHoldOn = nil
			}
		}

	}
	return
}

func importParse(reader *bufio.Reader, str string) (imports []*ImportStruct, inline bool) {
	if !strings.Contains(str, "(") {
		return []*ImportStruct{ImportStr(str)}, true
	}
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if strings.TrimSpace(contentStr) == ")" {
			break
		}
		imports = append(imports, ImportStr(contentStr))
	}
	return
}

func ImportStr(str string) *ImportStruct {
	if strings.HasPrefix(strings.TrimSpace(str), "import") {
		str = strings.TrimSpace(strings.ReplaceAll(str, "import", ""))
	}
	arr := strings.Split(str, "\"")
	return &ImportStruct{
		ImportTag:    strings.TrimSpace(arr[0]),
		ImportString: arr[1],
	}
}

func InlineImportInject(sourceStruct *SourceStruct, imports []*ImportStruct) {
	str := ""
	for _, one := range imports {
		if !contain(sourceStruct, one) {
			str += fmt.Sprint("import\t", one.ImportTag, " ", "\"", one.ImportString, "\"\n")
		}
	}
	str += "\n"
	err := util.InsertStringToFile(sourceStruct.Path, str, sourceStruct.ImportLine)
	if err != nil {
		log.Fatal(err)
	}
}

func MultiLineInject(sourceStruct *SourceStruct, imports []*ImportStruct) {
	str := ""
	for _, one := range imports {
		if !contain(sourceStruct, one) {
			str += fmt.Sprint("\t", one.ImportTag, " ", "\"", one.ImportString, "\"\n")
		}
	}
	str += "\n"
	err := util.InsertStringToFile(sourceStruct.Path, str, sourceStruct.ImportLine)
	if err != nil {
		log.Fatal(err)
	}
}

func contain(sourceStruct *SourceStruct, i *ImportStruct) bool {
	for _, one := range sourceStruct.Imports {
		if one.ImportString == i.ImportString {
			return true
		}
	}
	return false
}
