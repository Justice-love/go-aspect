package inject

import (
	"github.com/Justice-love/go-aspect/parse"
	"github.com/Justice-love/go-aspect/writer"
)

type TestInjectFile struct{}

func (i TestInjectFile) InjectFunc(sourceStruct *parse.SourceStruct, aspect *Aspect) {
	for _, one := range aspect.Point.imports {
		_ = parse.Contain(sourceStruct, one)
	}

	codeStruct := SourceStructStr(aspect.Function)
	writer.Append(sourceStruct.XgcPath, codeStruct)

	before := bindParam(aspect.Point.code+"\n", aspect)
	if aspect.Function.Receiver != nil {
		writer.ReplaceReceiver(sourceStruct.Path, aspect.Function.Receiver.Receiver, parse.FunctionReceiverLineNum(aspect.Function.LineString))
	} else {
		writer.AddReceiver(sourceStruct.Path, aspect.Function.FuncName, parse.FunctionNameLineNum(aspect.Function.LineString))
	}
	code := SourceFunctionStr(aspect.Function) + "\n"
	code += before
	if len(aspect.Function.Returns) > 0 {
		code += "\treturn " + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	} else {
		code += "\t" + aroundTargetWithNewReceiver(aspect.Function, aspect.Function.FuncName) + "\n"
	}
	code += "}"
	writer.Append(sourceStruct.XgcPath, code)
}

func (i TestInjectFile) Name() string {
	return "Test"
}
