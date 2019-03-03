package Test

import (
	. "InterpreterVM/Source/vm"
	"testing"
)

var gParser = newParserWrapper("")

func parse(s string) SyntaxTree {
	gParser.SetInput(s)
	return gParser.Parse()
}

func isEOF() bool {
	return gParser.IsEOF()
}

func TestParse1(t *testing.T) {
	root := parse("a = 1 + 2 + 3")
	expList := ASTFind(root, func(SyntaxTree) bool { return true }).(*ExpressionList)
	if len(expList.ExpList) != 1 {
		t.Error("parse1 error")
	}
	exp := expList.ExpList[0]
	binExp := exp.(*BinaryExpression)
	if binExp == nil {
		t.Error("parse1 error")
	}
	if binExp.OpToken.Token != '+' {
		t.Error("parse1 error")
	}
	if binExp.Right.(*Terminator) == nil {
		t.Error("parse1 error")
	}

	// 1 + 2
	if binExp == nil {
		t.Error("parse1 error")
	}
	if binExp.OpToken.Token != '+' {
		t.Error("parse1 error")
	}
	if binExp.Left.(*Terminator) == nil {
		t.Error("parse1 error")
	}
	if binExp.Right.(*Terminator) == nil {
		t.Error("parse1 error")
	}
}
