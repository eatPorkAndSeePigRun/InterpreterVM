package Test

import (
	"InterpreterVM/Source/vm"
	"container/list"
)

var gGC vm.GC
var (
	gGlobalTable    list.List
	gGlobalFunction list.List
	gGlobalClosure  list.List
	gGlobalString   list.List

	gScopeTable   list.List
	gScopeClosure list.List
	gScopeString  list.List
)
