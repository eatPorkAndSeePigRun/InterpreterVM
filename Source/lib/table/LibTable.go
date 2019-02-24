package table

import (
	. "InterpreterVM/Source/vm"
	"fmt"
)

// If the index value is number, then get the number value,
// else report type error.
func getNumber(api *StackAPI, index int, numType interface{}) bool {
	if !api.IsNumber(index) {
		api.ArgTypeError(index, ValueTNumber)
		return false
	}

	if num, ok := numType.(*int); ok {
		*num = int(api.GetNumber(index))
		return true
	}
	panic("assert")
}

func concat(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTTable) {
		return 0
	}

	table := api.GetTable(0)
	var sep string
	i := 1
	j := table.ArraySize()

	params := api.GetStackSize()
	if params > 1 {
		// If the value of index 1 is string, then get the string as sep
		if api.IsString(1) {
			sep = api.GetString(1).GetCStr()

			// Try to get the range of table
			if params > 2 && !getNumber(api, 2, &i) {
				return 0
			}

			if params > 3 && !getNumber(api, 3, &j) {
				return 0
			}
		} else {
			// Try to get the range of table
			if !getNumber(api, 1, &i) {
				return 0
			}

			if params > 2 && !getNumber(api, 2, &j) {
				return 0
			}
		}
	}

	key := NewValueNum(0.0)

	// Concat values(number or string) of the range [i, j]
	var str string
	for ; i <= j; i++ {
		key.Num = 1
		value := table.GetValue(key)

		if value.Type == ValueTNumber {
			str += fmt.Sprint(value.Num)
		} else {
			str += fmt.Sprint(value.Str.GetCStr())
		}

		if i != j {
			str += sep
		}
	}

	api.PushString(str)
	return 1
}

func insert(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(2, ValueTTable) {
		return 0
	}

	params := api.GetStackSize()
	table := api.GetTable(0)
	index := table.ArraySize() + 1
	value := 1

	if params > 2 {
		if !getNumber(api, 1, &index) {
			return 0
		}

		value = 2
	}

	api.PushBool(table.InsertArrayValue(index, *api.GetValue(value)))
	return 1
}

func pack(state *State) int {
	api := NewStackAPI(state)

	table := state.NewTable()
	params := api.GetStackSize()
	for i := 0; i < params; i++ {
		table.SetArrayValue(i+1, *api.GetValue(i))
	}

	api.PushTable(table)
	return 1
}

func remove(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(1, ValueTTable, ValueTNumber) {
		return 0
	}

	table := api.GetTable(0)
	index := table.ArraySize()

	// There is no elements in array of table.
	if index == 0 {
		api.PushBool(false)
		return 1
	}

	params := api.GetStackSize()
	if params > 1 {
		index = int(api.GetNumber(1))
	}

	api.PushBool(table.EraseArrayValue(index))
	return 1
}

func unpack(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(1, ValueTTable, ValueTNumber, ValueTNumber) {
		return 0
	}

	params := api.GetStackSize()
	table := api.GetTable(0)

	begin := 1
	end := table.ArraySize()
	if params > 1 {
		begin = int(api.GetNumber(1))
	}
	if params > 2 {
		end = int(api.GetNumber(2))
	}

	count := 0
	key := NewValueNum(0.0)
	for i := begin; i <= end; i++ {
		key.Num = float64(i)
		api.PushValue(table.GetValue(key))
		count++
	}

	return count
}

func RegisterLibTable(state *State) {
	lib := NewLibrary(state)
	table := [5]TableMemberReg{
		*NewTableMemberRegCFunction("concat", concat),
		*NewTableMemberRegCFunction("insert", insert),
		*NewTableMemberRegCFunction("pack", pack),
		*NewTableMemberRegCFunction("remove", remove),
		*NewTableMemberRegCFunction("unpack", unpack),
	}

	lib.RegisterTableFunction("table", &table[0], len(table))
}
