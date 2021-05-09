package parse

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"testing"
)

func TestSourceParse(t *testing.T) {
	r := parse("/Users/xuyi/go/src/eddy.org/go-aspect")
	fmt.Println(r)
}

func TestInlineFunction(t *testing.T) {
	r := funcInline("func TestSourceParse() {", 10)
	fmt.Println(r)
}

func parse(root string) (sources []*SourceStruct) {
	info, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for _, one := range info {
		if one.IsDir() {
			sources = append(sources, parse(fmt.Sprint(root, "/", one.Name()))...)
		} else {
			if strings.HasSuffix(one.Name(), ".go") && !strings.HasSuffix(one.Name(), "_test.go") {
				source := SourceParse(fmt.Sprint(root, "/", one.Name()))
				sources = append(sources, source)
			}
		}
	}
	return
}
