package Test

import (
	. "InterpreterVM/Source/vm"
	"testing"
)

func semantic(s string) SyntaxTree {
	gParser.SetInput(s)
	ast := gParser.Parse()
	SemanticAnalysis(ast, gParser.GetState())
	return ast
}

func TestSemantic1(t *testing.T) {

}
