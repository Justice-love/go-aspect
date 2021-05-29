package parse

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSourceParse(t *testing.T) {
	dir, _ := os.Getwd()
	r := parse(dir[:strings.LastIndex(dir, "/")])
	fmt.Println(SourcePrettyText(r))
}

func parse(root string) (sources []*SourceStruct) {
	info, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for _, one := range info {
		if one.IsDir() && one.Name() == "vendor" {
			continue
		}
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
