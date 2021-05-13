package util

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var regx = regexp.MustCompile(`\s+`)

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

func InsertStringToFile(path, str string, index int) error {
	lines, err := fileLines(path)
	if err != nil {
		return err
	}

	fileContent := ""
	for i, line := range lines {
		if i == index {
			fileContent += str
		}
		fileContent += line
		fileContent += "\n"
	}

	return ioutil.WriteFile(path, []byte(fileContent), 0644)
}

func SplitSpace(str string) []string {
	return regx.Split(strings.TrimSpace(str), -1)
}
