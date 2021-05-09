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
	log.Fatal("unsupported")
}

func BeforeInjectFile(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	ctxName := contextName(aspect.Function.Params)
	err := util.InsertStringToFile(sourceStruct.Path, bindParam(aspect.Point.code, ctxName)+"\n", aspect.Function.FuncLine)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

// import ""
//point before(package.owner.func) {
//	code here
//}
type Point struct {
	mode          InjectMod
	injectPackage string
	injectFunc    string
	injectOwn     string
	code          string
	imports       []*parse.ImportStruct
}

type Aspect struct {
	Function *parse.FuncStruct
	Point    *Point
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

func NewPoint(mode InjectMod, p, f, o string) *Point {
	if mode == nil {
		log.Fatalf("%s", "unsupported mode")
	}
	return &Point{
		mode:          mode,
		injectPackage: p,
		injectFunc:    f,
		injectOwn:     o,
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
		point = NewPoint(injectMap[fs[0]], ps[0], ps[1], "")
	} else {
		point = NewPoint(injectMap[fs[0]], ps[0], ps[2], ps[1])
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
			if f.FuncOwn && p.injectOwn == "" {
				continue
			}
			advice := NewAdvice(sourceStruct, f, p)
			advices = append(advices, advice)
			continue tag
		}
	}
	return
}
