package parse

import (
	"github.com/Justice-love/go-aspect/util"
	log "github.com/sirupsen/logrus"
	"strings"
)

type StructType func(str string) (param *ParamStruct)

var ParamTypes = map[string]StructType{
	"struct":    structFunc,
	"slice":     sliceFunc,
	"array":     arrayFunc,
	"map":       mapFunc,
	"func":      funcFunc,
	"interface": interfaceFunc,
}

func GetTypeStruct(t string) StructType {
	if ty, ok := ParamTypes[t]; ok {
		return ty
	} else {
		return structFunc
	}
}

func oneParam(str string) (paramStr string, remain string) {
	if len(str) == 0 {
		return
	}
	if strings.HasPrefix(str, ",") {
		str = strings.TrimSpace(str[1:])
	}
	if len(str) == 0 {
		return
	}
	index := strings.Index(str, ",")
	if index < 1 {
		return str, ""
	}
	sub := str[:index]
	if !strings.Contains(strings.TrimSpace(sub), " ") {
		return strings.TrimSpace(sub), str[index:]
	}
	pure := strings.TrimSpace(str)
	ton := 0
	for i, s := range pure {
		if ton == 0 && s == ',' {
			return pure[:i], pure[i:]
		} else if s == '(' || s == '{' || s == '[' {
			ton += 1
		} else if s == ')' || s == '}' || s == ']' {
			ton -= 1
		}
	}
	log.Fatalf("get param failer, %s", str)
	return
}

func hasType(str string) bool {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	return len(kvs) > 1
}

func chooseStructType(str string) StructType {
	t := strings.TrimSpace(str[strings.Index(str, " "):])
	if strings.HasPrefix(t, "func") {
		return funcFunc
	} else if strings.HasPrefix(t, "[]") {
		return sliceFunc
	} else if strings.HasPrefix(t, "[") {
		return arrayFunc
	} else if strings.HasPrefix(t, "map") {
		return mapFunc
	} else if strings.HasPrefix(t, "interface") {
		return interfaceFunc
	} else {
		return structFunc
	}
}

func structFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	types := strings.Split(kvs[1], ".")
	if len(types) == 1 {
		t, p := CheckPointer(types[0])
		param = &ParamStruct{Name: kvs[0], ParamType: t}
		param.Pointer = p
	} else {
		t, p := CheckPointer(types[1])
		param = &ParamStruct{Name: kvs[0], ParamType: t}
		param.Pointer = p
	}
	param.StructType = structFunc
	return
}

func sliceFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	param = &ParamStruct{
		Name:       kvs[0],
		ParamType:  "slice",
		StructType: sliceFunc,
	}
	return
}

func arrayFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	param = &ParamStruct{
		Name:       kvs[0],
		ParamType:  "array",
		StructType: arrayFunc,
	}
	return
}

func mapFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	param = &ParamStruct{
		Name:       kvs[0],
		ParamType:  "map",
		StructType: mapFunc,
	}
	return
}

func funcFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	param = &ParamStruct{
		Name:       kvs[0],
		ParamType:  "func",
		StructType: funcFunc,
	}
	return
}

func interfaceFunc(str string) (param *ParamStruct) {
	kvs := util.SplitSpace(strings.TrimSpace(str))
	param = &ParamStruct{
		Name:       kvs[0],
		ParamType:  "interface",
		StructType: interfaceFunc,
	}
	return
}
