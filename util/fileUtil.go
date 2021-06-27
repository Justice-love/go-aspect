package util

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var regx = regexp.MustCompile(`\s+`)
var Prefix = "xgc_"

func fileLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return linesFromReader(f)
}

func linesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func InsertStringToFile(path, str string, index int) {
	lines, err := fileLines(path)
	if err != nil {
		log.Fatal(err)
	}

	fileContent := ""
	for i, line := range lines {
		if i == index {
			fileContent += str
		}
		fileContent += line
		fileContent += "\n"
	}

	_ = ioutil.WriteFile(path, []byte(fileContent), 0644)
}

func SplitSpace(str string) []string {
	return regx.Split(strings.TrimSpace(str), -1)
}

func ReplaceFunctionName(path, old string, line int) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func() {
		_ = f.Close()
	}()
	reader := bufio.NewReader(f)
	l := 1
	buffer := bytes.Buffer{}
	for {
		content, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if l == line && !strings.Contains(contentStr, Prefix+old) {
			buffer.WriteString(strings.Replace(contentStr, old, Prefix+old, 1) + "\n")
		} else {
			buffer.WriteString(contentStr + "\n")
		}
		l += 1
	}
	_ = ioutil.WriteFile(path, buffer.Bytes(), 0644)
	return Prefix + old
}

func Append(path, code string) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	buffer := bytes.Buffer{}
	buffer.Write(bs)
	buffer.WriteString("\n\n")
	buffer.WriteString(code)
	_ = ioutil.WriteFile(path, buffer.Bytes(), 0644)
}

func InitFile(path, p string) {
	_ = ioutil.WriteFile(path, []byte("package "+p+"\n"), 0644)
}
