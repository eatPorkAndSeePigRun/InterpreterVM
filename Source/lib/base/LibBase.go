package base

import (
	. "InterpreterVM/Source/vm"
	"bufio"
	"fmt"
	"os"
)

func print(state *State) int {
	api := NewStackAPI(state)
	params := api.GetStackSize()

	for i := 0; i < params; i++ {
		vType := api.GetValueType(i)

		switch vType {
		case ValueTNil:
			fmt.Printf("nil")
		case ValueTBool:
			fmt.Printf("%t", api.GetBool(i))
		case ValueTNumber:
			fmt.Printf("%.14g", api.GetNumber(i))
		case ValueTString:
			fmt.Printf("%s", api.GetCString(i))
		case ValueTClosure:
			fmt.Printf("function:\t%p", api.GetClosure(i))
		case ValueTTable:
			fmt.Printf("table:\t%p", api.GetTable(i))
		case ValueTUserData:
			fmt.Printf("userdata:\t%p", api.GetUserData(i))
		case ValueTCFunction:
			fmt.Printf("function:\t%p", api.GetCFunction(i))
		default:
		}

		if i != params-1 {
			fmt.Printf("\t")
		}
	}

	fmt.Printf("\n")
	return 0
}

func puts(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	fmt.Printf("%s", api.GetCString(0))
	return 0
}

func dataType(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs1(1) {
		return 0
	}

	v := api.GetValue(0)
	var vType int
	if v.Type == ValueTUpvalue {
		vType = v.Upvalue.GetValue().Type
	} else {
		vType = v.Type
	}

	switch vType {
	case ValueTNil:
		api.PushString("nil")
	case ValueTBool:
		api.PushString("boolean")
	case ValueTNumber:
		api.PushString("number")
	case ValueTString:
		api.PushString("string")
	case ValueTTable:
		api.PushString("table")
	case ValueTUserData:
		api.PushString("userdata")
	case ValueTClosure, ValueTCFunction:
		api.PushString("function")
	default:
		panic("assert")
		return 0
	}
	return 1
}

func doIPairs(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs1(2, ValueTTable, ValueTNumber) {
		return 0
	}

	t := api.GetTable(0)
	num := api.GetNumber(1) + 1

	k := NewValueNum(num)
	v := t.GetValue(k)

	if v.Type == ValueTNil {
		return 0
	}

	api.PushValue(k)
	api.PushValue(v)
	return 2
}

func iPairs(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTTable) {
		return 0
	}

	t := api.GetTable(0)
	api.PushCFunction(doIPairs)
	api.PushTable(t)
	api.PushNumber(0)
	return 3
}

func doPairs(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(2, ValueTTable) {
		return 0
	}

	t := api.GetTable(0)
	lastKey := api.GetValue(1)

	var key, value Value
	if lastKey.Type == ValueTNil {
		t.FirstKeyValue(&key, &value)
	} else {
		t.NextKeyValue(lastKey, &key, &value)
	}

	api.PushValue(key)
	api.PushValue(value)
	return 2
}

func pairs(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTTable) {
		return 0
	}

	t := api.GetTable(0)
	api.PushCFunction(doPairs)
	api.PushTable(t)
	api.PushNil()
	return 3
}

func getLine(state *State) int {
	api := NewStackAPI(state)
	line := fmt.Sprint(bufio.NewScanner(os.Stdin).Text())
	api.PushString(line)
	return 1
}

func require(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	module := api.GetString(0).GetStdString()
	if !state.IsModuleLoaded(module) {
		state.DoModule(module)
	}

	return 0
}

func RegisterLibBase(state *State) {
	lib := NewLibrary(state)
	lib.RegisterFunc("print", print)
	lib.RegisterFunc("puts", puts)
	lib.RegisterFunc("ipairs", iPairs)
	lib.RegisterFunc("pairs", pairs)
	lib.RegisterFunc("type", dataType)
	lib.RegisterFunc("getline", getLine)
	lib.RegisterFunc("require", require)
}
