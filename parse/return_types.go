package parse

import (
	"regexp"
	"strings"
)

const (
	none ReturnType = iota
	StructReturn
	SliceReturn
	ArrayReturn
	MapReturn
	FuncReturn
	InterFaceReturn
	ChanReturn
)

var regx, _ = regexp.Compile("\\s+")

type ReturnType int

var ReturnNameMap = map[ReturnType]string{
	StructReturn:    "struct",
	SliceReturn:     "slice",
	ArrayReturn:     "array",
	MapReturn:       "map",
	FuncReturn:      "func",
	InterFaceReturn: "interface",
	ChanReturn:      "chan",
}

func oneReturn(str string) (returnStr, remain string) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return
	}
	count := 0
	for i, c := range str {
		if count == 0 && c == ',' {
			returnStr = str[:i]
			remain = str[i+1:]
			break
		} else if count == 0 && i == len(str)-1 {
			returnStr = str
			break
		} else if c == '(' || c == '{' || c == '[' {
			count += 1
		} else if c == ')' || c == '}' || c == ']' {
			count -= 1
		}
	}
	return
}

func chooseReturnType(str string) (ReturnType, bool) {
	str = strings.Trim(str, "->")
	str = strings.Trim(str, "<-")
	str = strings.TrimSpace(str)
	if !strings.Contains(str, " ") {
		return choose(str), false
	}
	if strings.HasPrefix(str, "func") {
		return FuncReturn, false
	}
	arr := regx.Split(str, 2)
	return choose(strings.TrimSpace(arr[1])), true
}

func choose(str string) ReturnType {
	if strings.HasPrefix(str, "[]") {
		return SliceReturn
	} else if strings.HasPrefix(str, "[") {
		return ArrayReturn
	} else if strings.HasPrefix(str, "map") {
		return MapReturn
	} else if strings.HasPrefix(str, "interface") {
		return InterFaceReturn
	} else if strings.HasPrefix(str, "chan") {
		return ChanReturn
	} else if strings.HasPrefix(str, "func") {
		return FuncReturn
	} else {
		return none
	}
}
