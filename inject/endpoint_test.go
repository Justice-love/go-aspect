package inject

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestEndpoints(t *testing.T) {
	arr := endpoints("../testdata/aspect_around.point")
	fmt.Println(EndpointPrettyText(arr))
}

func TestOnePoint(t *testing.T) {
	a := assert.New(t)
	point := "\tsample.AfterPrint({{c}})\n}"
	reader := bufio.NewReader(strings.NewReader(point))
	p := onePoint("point after(main.Do(c Context)) {", reader)
	a.NotNil(p)
	a.Equal(&AfterInjectFile{}, p.mode)
}

func TestOneOtherPoint(t *testing.T) {
	point := "\tsample.AfterPrint({{c}})\n}"
	reader := bufio.NewReader(strings.NewReader(point))
	_ = onePoint("point other(main.Do(c Context)) {", reader)
}
