package parse

import (
	"bufio"
	"eddy.org/go-aspect/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sort"
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
	Receiver    *ReceiverStruct
	FuncName    string
	FuncLine    int
	FuncEndLine int
	Params      []*ParamStruct
}

type ReceiverStruct struct {
	Pointer  bool
	Alias    string
	Receiver string
}

type ParamStruct struct {
	Pointer    bool
	Name       string
	ParamType  string
	StructType StructType
}

func NewFuncStruct(name string, line *int) *FuncStruct {
	return &FuncStruct{
		FuncName: name,
		FuncLine: *line,
	}
}

func NewSourceStruct(path string) *SourceStruct {
	return &SourceStruct{Path: path, InjectImport: MultiLineInject}
}

type ParamSort []*ParamStruct

func (p ParamSort) Len() int {
	return len(p)
}

func (p ParamSort) Less(i, j int) bool {
	return fmt.Sprint(p[i].Name, "_", p[i].ParamType) > fmt.Sprint(p[j].Name, "_", p[j].ParamType)
}

func (p ParamSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
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
			funcLine := &line
			funcS := funcParse(reader, contentStr, funcLine)
			funcEndParse(reader, funcLine)
			funcS.FuncEndLine = *funcLine - 1
			line = *funcLine
			source.Funcs = append(source.Funcs, funcS)
		}
	}
	if source.Imports == nil {
		source.ImportLine = 2
		source.InjectImport = InlineImportInject
	}
	return source
}

func funcEndParse(reader *bufio.Reader, line *int) {
	count := 1
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		*line += 1
		contentStr := string(content)
		for _, r := range contentStr {
			if r == '}' {
				count -= 1
			} else if r == '{' {
				count += 1
			}
		}
		if count == 0 {
			break
		}
	}
}

func funcParse(reader *bufio.Reader, str string, line *int) *FuncStruct {
	if strings.HasSuffix(strings.TrimSpace(str), "{") {
		return funcInline(str, line)
	} else {
		return funcMultiLine(reader, str, line)
	}
}

func funcMultiLine(reader *bufio.Reader, str string, line *int) *FuncStruct {
	funcStr := str
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		*line += 1
		contentStr := string(content)
		funcStr += contentStr
		if strings.HasSuffix(strings.TrimSpace(contentStr), "{") {
			break
		}
	}
	return funcInline(funcStr, line)
}

//TODO: skip "}"
func funcInline(str string, line *int) (fun *FuncStruct) {
	str = strings.TrimSpace(strings.TrimLeft(str, "func"))
	var (
		receiver *ReceiverStruct
	)
	if strings.HasPrefix(str, "(") {
		receiver = inlineReceiver(str[1:strings.Index(str, ")")])
		str = strings.TrimSpace(str[strings.Index(str, ")")+1:])
	}
	funName := str[:strings.Index(str, "(")]
	paramStr := paramStr(str[strings.Index(str, "("):])
	params := inlineParam(paramStr)
	fun = NewFuncStruct(funName, line)
	fun.Receiver = receiver
	fun.Params = params
	return
}

func paramStr(str string) string {
	if len(str) == 0 {
		return str
	}
	count := 0
	for i, s := range str {
		if i > 0 && count == 0 {
			return str[1 : i-1]
		}
		if s == '(' {
			count += 1
		} else if s == ')' {
			count -= 1
		}
	}
	log.Fatalf("get whold param string failer, %v", str)
	return ""
}

func inlineReceiver(s string) *ReceiverStruct {
	var r *ReceiverStruct
	arr := util.SplitSpace(strings.TrimSpace(s))
	if len(arr) == 1 {
		_, p := CheckPointer(strings.TrimSpace(arr[0]))
		r = &ReceiverStruct{
			Pointer:  p,
			Alias:    "",
			Receiver: receiveType(arr[0]),
		}
	} else {
		_, p := CheckPointer(strings.TrimSpace(arr[1]))
		r = &ReceiverStruct{
			Pointer:  p,
			Alias:    strings.TrimSpace(arr[0]),
			Receiver: receiveType(arr[1]),
		}
	}
	return r
}

func receiveType(s string) string {
	arr := strings.Split(strings.TrimLeft(strings.TrimSpace(s), "*"), ".")
	return arr[len(arr)-1]
}

func CheckPointer(s string) (t string, p bool) {
	if strings.HasPrefix(s, "*") {
		t = s[1:]
		p = true
	} else {
		t = s
	}
	return
}

func inlineParam(str string) (params []*ParamStruct) {
	var typeHoldOn []*ParamStruct
	for {
		var paramStr string
		paramStr, str = oneParam(str)
		if len(paramStr) == 0 {
			break
		}
		if !hasType(paramStr) {
			typeHoldOn = append(typeHoldOn, &ParamStruct{Name: strings.TrimSpace(paramStr)})
		} else {
			param := chooseStructType(paramStr)(paramStr)
			if len(typeHoldOn) != 0 {
				for _, one := range typeHoldOn {
					one.ParamType = param.ParamType
					one.StructType = param.StructType
					one.Pointer = param.Pointer
					params = append(params, one)
				}
				typeHoldOn = nil
			}
			params = append(params, param)
		}
	}
	sort.Sort(ParamSort(params))
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
