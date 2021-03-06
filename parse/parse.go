package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Justice-love/go-aspect/writer"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type SourceStruct struct {
	Path         string
	XgcPath      string
	PackageStr   string
	Imports      []*ImportStruct
	Funcs        []*FuncStruct
	InjectImport func(sourceStruct *SourceStruct, imports []*ImportStruct)
	ImportLine   int
	FileLine     int
}

type ImportStruct struct {
	ImportTag     string
	ImportString  string
	ImportEndTerm string
	SourceTag     string
	SourceContain bool
}

type FuncStruct struct {
	FuncString  string
	Receiver    *ReceiverStruct
	FuncName    string
	FuncLine    int
	FuncEndLine int
	Params      []*ParamStruct
	ReturnLine  int
	Returns     []ReturnType
	NameLine    int
	LineString  []*FuncLine
}

type FuncLine struct {
	LineString string
	LineNum    int
	Types      int
}

const (
	ReceiverLine     = 1 << 1
	FunctionNameLine = 1 << 2
)

type ReceiverStruct struct {
	Pointer  bool
	Alias    string
	Receiver string
}

type ParamStruct struct {
	Pointer    bool
	Name       string
	ParamType  string
	StructType ParamStructType
}

func NewFuncStruct(name string, line *int) *FuncStruct {
	return &FuncStruct{
		FuncName: name,
		FuncLine: *line,
	}
}

func NewSourceStruct(path string) *SourceStruct {
	xgc_path := fmt.Sprint(strings.TrimRight(path, ".go"), "_xgc.go")
	return &SourceStruct{Path: path, XgcPath: xgc_path}
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
			source.FileLine = line
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
			} else {
				line += len(imr) + 1
				source.ImportLine = line - 1
			}
		} else if strings.HasPrefix(contentStr, "func") {
			funcLine := &line
			funcS := funcParse(reader, contentStr, funcLine)
			if returnLine := funcEndParse(reader, funcLine); returnLine > 0 {
				funcS.ReturnLine = returnLine
			}
			funcS.FuncEndLine = *funcLine - 1
			line = *funcLine
			source.Funcs = append(source.Funcs, funcS)
		}
	}
	if source.Imports == nil {
		source.ImportLine = 2
	}
	return source
}

func funcEndParse(reader *bufio.Reader, line *int) (returnLine int) {
	skipHolder := &SkipHolder{}
	count := 1
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		*line += 1
		contentStr := string(content)
		if strings.HasPrefix(strings.TrimSpace(contentStr), "return ") || strings.TrimSpace(contentStr) == "return" {
			returnLine = *line
		}
		for _, r := range contentStr {
			if skipHolder.NeedSkip(r) {
				continue
			}
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
	return
}

func funcParse(reader *bufio.Reader, str string, line *int) (f *FuncStruct) {
	if strings.HasSuffix(strings.TrimSpace(str), "{") {
		f = funcInline(str, line)
		f.FuncString = str
	} else {
		f = funcMultiLine(reader, str, line)
	}
	return
}

func funcMultiLine(reader *bufio.Reader, str string, line *int) *FuncStruct {
	fls := make([]*FuncLine, 0)
	funcStr := str
	fls = append(fls, &FuncLine{LineString: str, LineNum: *line})
	fs := make(map[int]string)
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		*line += 1
		contentStr := string(content)
		fs[*line] = contentStr
		funcStr += contentStr
		fls = append(fls, &FuncLine{LineString: contentStr, LineNum: *line})
		if strings.HasSuffix(strings.TrimSpace(contentStr), "{") {
			break
		}
	}
	f := funcInline(funcStr, line)
	f.FuncString = funcStr
	f.LineString = fls
	for k, v := range fs {
		if strings.Contains(v, f.FuncName) {
			f.NameLine = k
			break
		}
	}
	for _, l := range fls {
		if strings.Contains(l.LineString, f.FuncName) {
			l.Types = l.Types | FunctionNameLine
		}
		if f.Receiver != nil && strings.Contains(l.LineString, f.Receiver.Receiver) {
			l.Types = l.Types | ReceiverLine
		}
	}
	return f
}

func funcInline(str string, line *int) (fun *FuncStruct) {
	basic := str
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
	returnStr := returnStr(str[strings.Index(str, "("):], paramStr)
	params := inlineParam(paramStr)
	returns := inlineReturns(returnStr)
	fun = NewFuncStruct(funName, line)
	fun.Receiver = receiver
	fun.Params = params
	fun.Returns = returns
	fun.NameLine = *line
	fun.LineString = []*FuncLine{{
		LineString: basic,
		LineNum:    *line,
		Types:      0 | FunctionNameLine | ReceiverLine,
	}}
	return
}

func inlineReturns(str string) []ReturnType {
	var returns []ReturnType
	var types []bool
	var returnStr string
	for {
		returnStr, str = oneReturn(str)
		if len(returnStr) == 0 {
			break
		}
		t, c := chooseReturnType(returnStr)
		returns = append(returns, t)
		types = append(types, c)

	}
	var check []int
	for i, c := range types {
		if !c {
			check = append(check, i)
		}
		if c && len(check) > 0 {
			for _, j := range check {
				returns[j] = returns[i]
			}
			check = nil
		}
	}
	for i, o := range returns {
		if o == none {
			returns[i] = StructReturn
		}
	}
	return returns
}

func returnStr(s string, str string) string {
	remain := s[len(str)+1:]
	remain = strings.Replace(remain, ")", "", 1)
	remain = strings.TrimSpace(strings.TrimRight(remain, "{"))
	if strings.HasPrefix(remain, "(") {
		return remain[1 : len(remain)-1]
	} else {
		return remain
	}
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
	arr := writer.SplitSpace(strings.TrimSpace(s))
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
		if strings.TrimSpace(contentStr) == "" {
			continue
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
		ImportTag:     strings.TrimSpace(arr[0]),
		ImportString:  arr[1],
		ImportEndTerm: endTerm(arr[1]),
	}
}

func endTerm(s string) string {
	arr := strings.Split(s, "/")
	return arr[len(arr)-1]
}

func Contain(sourceStruct *SourceStruct, i *ImportStruct) bool {
	for _, one := range sourceStruct.Imports {
		if one.ImportString == i.ImportString {
			i.SourceTag = one.ImportTag
			i.SourceContain = true
			return true
		}
	}
	return false
}

func SourcePrettyText(
	sources []*SourceStruct) string {
	var buff bytes.Buffer
	for j, source := range sources {
		_, _ = buff.WriteString("{\n")
		_, _ = buff.WriteString("\tpath: " + source.Path)
		_, _ = buff.WriteString("\n")
		_, _ = buff.WriteString("\tpackage ")
		_, _ = buff.WriteString("\x1b[31m")
		_, _ = buff.WriteString("" + source.PackageStr)
		_, _ = buff.WriteString("\x1b[0m")
		_, _ = buff.WriteString("\n")
		if len(source.Imports) > 0 {
			_, _ = buff.WriteString("\timport {\n")
		}
		for _, im := range source.Imports {
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString("\t\t" + im.ImportString)
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString("\n")
		}
		if len(source.Imports) > 0 {
			_, _ = buff.WriteString("\t}\n")
		}
		_, _ = buff.WriteString("\n")
		for _, fu := range source.Funcs {
			_, _ = buff.WriteString("\tfunc ")
			if fu.Receiver != nil {
				_, _ = buff.WriteString("(\x1b[31m")
				if fu.Receiver.Pointer {
					_, _ = buff.WriteString("*")
				}
				_, _ = buff.WriteString(fu.Receiver.Receiver)
				_, _ = buff.WriteString("\x1b[0m) ")
			}
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(fu.FuncName)
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString("(")
			for i, p := range fu.Params {
				_, _ = buff.WriteString("\x1b[32m")
				_, _ = buff.WriteString(p.Name)
				_, _ = buff.WriteString(" ")
				_, _ = buff.WriteString(p.ParamType)
				_, _ = buff.WriteString("\x1b[0m")
				if i < len(fu.Params)-1 {
					_, _ = buff.WriteString(", ")
				}
			}
			_, _ = buff.WriteString(") returns (")
			for i, p := range fu.Returns {
				_, _ = buff.WriteString("\x1b[32m")
				_, _ = buff.WriteString(ReturnNameMap[p])
				_, _ = buff.WriteString("\x1b[0m")
				if i < len(fu.Returns)-1 {
					_, _ = buff.WriteString(", ")
				}
			}
			_, _ = buff.WriteString(")\n")
		}
		if j < len(sources)-1 {
			_, _ = buff.WriteString("},\n")
		} else {
			_, _ = buff.WriteString("}\n")
		}
	}
	return buff.String()
}

func FunctionNameLineNum(lines []*FuncLine) int {
	for _, l := range lines {
		if l.Types&FunctionNameLine == FunctionNameLine {
			return l.LineNum
		}
	}
	return 0
}

func FunctionReceiverLineNum(lines []*FuncLine) int {
	for _, l := range lines {
		if l.Types&ReceiverLine == ReceiverLine {
			return l.LineNum
		}
	}
	return 0
}
