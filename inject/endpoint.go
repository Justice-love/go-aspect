package inject

import (
	"bufio"
	"bytes"
	"github.com/Justice-love/go-aspect/parse"
	"github.com/Justice-love/go-aspect/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"reflect"
	"strings"
)

type Point struct {
	mode           CodeInjectInterface
	injectPackage  string
	injectFunc     string
	injectReceiver *EndpointReceiver
	code           string
	imports        []*parse.ImportStruct
	params         []*parse.ParamStruct
}

type Aspect struct {
	Function *parse.FuncStruct
	Point    *Point
}

type EndpointReceiver struct {
	Pointer  bool
	Receiver string
}

type Advice struct {
	Source *parse.SourceStruct
	Aspect []*Aspect
	fi     []*FunctionInject
}

func NewAdvice(source *parse.SourceStruct, function *parse.FuncStruct, point *Point) *Advice {
	return &Advice{
		Source: source,
		Aspect: []*Aspect{
			{
				Function: function,
				Point:    point,
			},
		},
	}
}

func NewPoint(mode CodeInjectInterface, p string) *Point {
	if mode == nil {
		log.Fatalf("%s", "unsupported mode")
	}
	return &Point{
		mode:          mode,
		injectPackage: p,
	}
}

func AllEndpoints(path []string) (points []*Point) {
	for _, one := range path {
		points = append(points, endpoints(one)...)
	}
	return
}

func endpoints(endpointPath string) (points []*Point) {
	f, err := os.Open(endpointPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func() {
		_ = f.Close()
	}()
	reader := bufio.NewReader(f)
	imports := make([]*parse.ImportStruct, 0)
	for {
		content, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "point") {
			point := onePoint(contentStr, reader)
			point.imports = parse.ImportCopy(imports)
			points = append(points, point)
		} else if strings.HasPrefix(contentStr, "import") {
			imports = append(imports, parse.ImportStr(contentStr))
		}
	}
	return
}

func onePoint(str string, reader *bufio.Reader) (point *Point) {
	pure := strings.TrimSpace(strings.TrimLeft(str, "point"))
	fs := strings.FieldsFunc(pure, func(r rune) bool {
		return r == '(' || r == ')'
	})
	if len(fs) != 3 && len(fs) != 4 {
		log.Fatalf("bad point %s", str)
		return
	}
	funs := strings.Split(fs[1], ".")
	point = NewPoint(injectMap[fs[0]], funs[0])
	if len(funs) == 2 {
		point.injectFunc = funs[1]
	} else {
		point.injectFunc = funs[2]
		point.injectReceiver = endpointReceiver(funs[1])
	}
	if len(fs) == 4 {
		point.params = endpointParams(fs[2])
	}
	point.code = endpointCode(reader)
	return
}

func endpointCode(reader *bufio.Reader) (code string) {
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if strings.HasPrefix(strings.TrimSpace(contentStr), "}") {
			break
		}
		code += contentStr + "\n"
	}
	return
}

func endpointParams(s string) (params []*parse.ParamStruct) {
	paramStrs := strings.Split(s, ",")
	for _, one := range paramStrs {
		nameAndType := util.SplitSpace(one)
		t, p := parse.CheckPointer(nameAndType[1])
		typeFunc := parse.GetTypeStruct(t)
		params = append(params, &parse.ParamStruct{Pointer: p, Name: nameAndType[0], StructType: typeFunc, ParamType: t})
	}
	return
}

func endpointReceiver(s string) *EndpointReceiver {
	t, p := parse.CheckPointer(s)
	return &EndpointReceiver{
		Pointer:  p,
		Receiver: t,
	}
}

func Match(sourceStruct *parse.SourceStruct, points []*Point) (advices []*Advice) {
tag:
	for _, p := range points {
		if p.injectPackage != sourceStruct.PackageStr {
			return nil
		}
	tag2:
		for _, f := range sourceStruct.Funcs {
			if f.FuncName != p.injectFunc {
				continue
			}
			if p.injectReceiver == nil && f.Receiver != nil {
				continue
			}
			if p.injectReceiver != nil && f.Receiver == nil {
				continue
			}
			if f.Receiver != nil && p.injectReceiver.Receiver != f.Receiver.Receiver {
				continue
			}
			if f.Receiver != nil && p.injectReceiver.Receiver == f.Receiver.Receiver && p.injectReceiver.Pointer != f.Receiver.Pointer {
				continue
			}
			if len(f.Params) != len(p.params) {
				continue
			}
			for i, param := range f.Params {
				pp := p.params[i]
				paramFunc := reflect.ValueOf(param.StructType)
				pointParamFunc := reflect.ValueOf(pp.StructType)
				if param.ParamType != pp.ParamType || paramFunc.Pointer() != pointParamFunc.Pointer() || param.Pointer != pp.Pointer {
					continue tag2
				}
			}
			advice := NewAdvice(sourceStruct, f, p)
			advices = append(advices, advice)
			continue tag
		}
	}
	return
}

func EndpointPrettyText(endpoints []*Point) string {
	var buff bytes.Buffer
	for _, one := range endpoints {
		_, _ = buff.WriteString(one.mode.Name())
		_, _ = buff.WriteString("\t")
		if one.injectReceiver != nil {
			_, _ = buff.WriteString("(")
			if one.injectReceiver.Pointer {
				_, _ = buff.WriteString("*")
			}
			_, _ = buff.WriteString(one.injectReceiver.Receiver)
			_, _ = buff.WriteString(")")
			_, _ = buff.WriteString("\t")
		}
		_, _ = buff.WriteString(one.injectPackage)
		_, _ = buff.WriteString(".")
		_, _ = buff.WriteString(one.injectFunc)
		_, _ = buff.WriteString("(")
		for i, p := range one.params {
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(p.Name)
			_, _ = buff.WriteString(" ")
			_, _ = buff.WriteString(p.ParamType)
			_, _ = buff.WriteString("\x1b[0m")
			if i < len(one.params)-1 {
				_, _ = buff.WriteString(", ")
			}
		}
		_, _ = buff.WriteString(")")
		_, _ = buff.WriteString("\n")
	}
	return buff.String()
}
