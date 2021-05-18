package file

import (
	"fmt"
	copy2 "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

func DebugDir() string {
	dest, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("%v", err)
	}
	dest = fmt.Sprint(dest, "/", ".xgc")
	err = os.MkdirAll(dest, 0700)
	if err != nil {
		log.Fatal(err)
	}
	return dest
}

func TempDir() string {
	dest, err := ioutil.TempDir("", "go-aspect")
	if err != nil {
		log.Fatalf("%v", err)
	}
	return dest
}

func SourceCopy(temp, source string) string {
	if temp == "" {
		log.Fatalf("error dest")
	}
	dest := strings.Join([]string{temp, source}, "/")
	_ = os.MkdirAll(dest, 0777)
	path, _ := os.Getwd()
	if err := copy2.Copy(path, dest); err != nil {
		log.Fatal(err)
	}
	return dest
}

func SourceDir() string {
	path, _ := os.Getwd()
	index := strings.LastIndex(path, "/")
	return path[index+1:]
}
