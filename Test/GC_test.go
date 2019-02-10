package Test

import (
	"InterpreterVM/Source/luna"
	"container/list"
)

var gGC luna.GC
var (
	gGlobalTable    list.List
	gGlobalFunction list.List
	gGlobalClosure  list.List
	gGlobalString   list.List

	gScopeTable   list.List
	gScopeClosure list.List
	gScopeString  list.List
)
