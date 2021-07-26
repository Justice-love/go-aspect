package inject

import (
	log "github.com/sirupsen/logrus"
	"regexp"
)

func funcNameRegx(regx string) *regexp.Regexp {
	r, e := regexp.Compile(regx)
	if e != nil {
		log.Errorf("error build regx, %s, %v", regx, e)
		return nil
	}
	return r
}
