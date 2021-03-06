package file

import (
	"fmt"
	"github.com/Justice-love/go-aspect/inject"
	"github.com/Justice-love/go-aspect/parse"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

type X struct {
	RootPath string
	Points   []*inject.Point
}

func (x *X) IteratorSource(root string) (injects []*inject.Advice) {
	info, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for _, one := range info {
		if one.IsDir() && one.Name() == "vendor" {
			continue
		}
		if one.IsDir() {
			injects = append(injects, x.IteratorSource(fmt.Sprint(root, "/", one.Name()))...)
		} else {
			if strings.HasSuffix(one.Name(), ".go") && !strings.HasSuffix(one.Name(), "_test.go") {
				source := parse.SourceParse(fmt.Sprint(root, "/", one.Name()))
				if advice := inject.Match(source, x.Points); len(advice) > 0 {
					injects = append(injects, advice...)
				}
			}
		}
	}
	return
}
