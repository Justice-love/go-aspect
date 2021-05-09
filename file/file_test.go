package file

import (
	"eddy.org/go-aspect/parse"
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestSourceDir(t *testing.T) {
	fmt.Println(SourceDir())
}

func TestParse(t *testing.T) {
	str := "func (s) TestParse() {"
	str = strings.TrimSpace(strings.TrimLeft(str, "func"))
	s := strings.FieldsFunc(str, func(r rune) bool {
		return r == '(' || r == ')'
	})
	fmt.Println(s)
}

func TestP(t *testing.T) {
	str := "\"fmt\""
	a := strings.Split(str, "\"")
	fmt.Println(a)
	vs := make([]*parse.FuncStruct, 2, 2)
	vs[0] = &parse.FuncStruct{FuncLine: 1}
	vs[1] = &parse.FuncStruct{FuncLine: 2}
	sort.SliceStable(vs, func(i, j int) bool {
		return vs[i].FuncLine > vs[j].FuncLine
	})
	fmt.Println(vs)
}
