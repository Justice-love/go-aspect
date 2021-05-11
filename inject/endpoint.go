package inject

import (
	"bufio"
	"eddy.org/go-aspect/parse"
	"eddy.org/go-aspect/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

var injectMap = map[string]InjectMod{
	"before": BeforeInjectFile,
	"after":  AfterInjectFile,
}

type InjectMod func(sourceStruct *parse.SourceStruct, aspect *Aspect)

func AfterInjectFile(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	code := `
	defer func() {
		` + aspect.Point.code + `
	}()` + "\n"

	ctxName := contextName(aspect.Function.Params)
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(code, ctxName), aspect.Function.FuncLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func BeforeInjectFile(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	ctxName := contextName(aspect.Function.Params)
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code, ctxName)+"\n", aspect.Function.FuncLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

type Point struct {
	mode           InjectMod
	injectPackage  string
	injectFunc     string
	injectReceiver *EndpointReceiver
	code           string
	imports        []*parse.ImportStruct
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

func NewPoint(mode InjectMod, p, f string) *Point {
	if mode == nil {
		log.Fatalf("%s", "unsupported mode")
	}
	return &Point{
		mode:          mode,
		injectPackage: p,
		injectFunc:    f,
	}
}

func Endpoints(root string) (points []*Point) {
	f, err := os.Open(fmt.Sprint(root, "/", "aspect.point"))
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
			point.imports = imports
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
	if len(fs) != 3 {
		log.Fatal("unsupported")
	}
	ps := strings.Split(fs[1], ".")
	if len(ps) == 2 {
		point = NewPoint(injectMap[fs[0]], ps[0], ps[1])
	} else {
		point = NewPoint(injectMap[fs[0]], ps[0], ps[2])
		var receiver *EndpointReceiver
		if strings.HasPrefix(ps[1], "*") {
			receiver = &EndpointReceiver{
				Pointer:  true,
				Receiver: ps[1][1:],
			}
		} else {
			receiver = &EndpointReceiver{
				Pointer:  false,
				Receiver: ps[1],
			}
		}
		point.injectReceiver = receiver
	}
	for {
		content, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if strings.HasPrefix(strings.TrimSpace(contentStr), "}") {
			break
		}
		point.code += contentStr
	}
	return point
}

func Match(sourceStruct *parse.SourceStruct, points []*Point) (advices []*Advice) {
tag:
	for _, p := range points {
		if p.injectPackage != sourceStruct.PackageStr {
			return nil
		}
		for _, f := range sourceStruct.Funcs {
			if !f.Context {
				continue
			}
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
			advice := NewAdvice(sourceStruct, f, p)
			advices = append(advices, advice)
			continue tag
		}
	}
	return
}
