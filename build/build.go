package build

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

var (
	BuildA bool // -a flag
	//BuildBuildmode         string // -buildmode flag
	BuildI          bool   // -i flag
	BuildLinkshared bool   // -linkshared flag
	BuildMSan       bool   // -msan flag
	BuildN          bool   // -n flag
	BuildO          string // -o flag
	//BuildPkgdir            string             // -pkgdir flag
	//BuildRace              bool               // -race flag
	//BuildTrimpath          bool // -trimpath flag
	//BuildV                 bool // -v flag
	//BuildWork              bool // -work flag
	//BuildX                 bool // -x flag

)

func Clean(root string) {
	_ = os.RemoveAll(root)
}

func Build(folder string, flags string, args string) {
	log.Debugf("flags:%s", flags)
	var cmd *exec.Cmd
	if strings.TrimSpace(args) == "" {
		cmd = exec.Command("/bin/bash", "-c", "go build "+flags+" .")
	} else {
		cmd = exec.Command("/bin/bash", "-c", "go build "+flags+" "+args)
	}
	cmd.Dir = folder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	_ = cmd.Wait()
}

func BuildTags(path string) string {
	tags := ""
	if BuildA {
		tags += " -a"
	}
	if BuildI {
		tags += " -i"
	}
	if BuildLinkshared {
		tags += " -linkshared true"
	}
	if BuildMSan {
		tags += " -msan"
	}
	if BuildN {
		tags += " -n"
	}
	if BuildO != "" {
		tags += " -o " + BuildO
	} else {
		tags += " -o " + path
	}
	return tags
}
