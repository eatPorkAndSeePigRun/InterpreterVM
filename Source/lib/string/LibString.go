package string

import (
	. "InterpreterVM/Source/vm"
	"math"
	"strings"
)

func abyte(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(1, ValueTString, ValueTNumber, ValueTNumber) {
		return 0
	}

	params := api.GetStackSize()

	i := 1
	if params >= 2 {
		i = int(api.GetNumber(1))
	}

	j := i
	if params >= 3 {
		j = int(api.GetNumber(2))
	}

	if i <= 0 {
		return 0
	}

	str := api.GetString(0)
	s := str.GetCStr()
	len := str.GetLength()
	count := 0

	for index := i - 1; index < j; index++ {
		if index >= 0 && index < len {
			api.PushNumber(float64(s[index]))
			count++
		}
	}
	return count
}

func char(state *State) int {
	api := NewStackAPI(state)
	params := api.GetStackSize()

	var str string
	for i := 0; i < params; i++ {
		if !api.IsNumber(i) {
			api.ArgTypeError(i, ValueTNumber)
			return 0
		} else {
			str += string(int(api.GetNumber(i)))
		}
	}

	api.PushString(str)
	return 1
}

func alen(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	api.PushNumber(float64(api.GetString(0).GetLength()))
	return 1
}

func lower(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	str := api.GetString(0)
	size := str.GetLength()
	cStr := str.GetCStr()

	lower := ""
	for i := 0; i < size; i++ {
		lower += strings.ToLower(string(cStr[i]))
	}

	api.PushString(lower)
	return 1
}

func reverse(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	str := api.GetString(0)
	size := str.GetLength()
	cStr := str.GetCStr()

	reverse := ""
	for ; size > 0; size-- {
		reverse += string(cStr[size-1])
	}

	api.PushString(reverse)
	return 1
}

func sub(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(2, ValueTString, ValueTNumber, ValueTNumber) {
		return 0
	}

	str := api.GetString(0)
	size := str.GetLength()
	cStr := str.GetCStr()
	start := int(api.GetNumber(1))
	end := size

	params := api.GetStackSize()
	if params <= 2 {
		if start == 0 {
			start = 1
		} else if start < 0 {
			start += size
			if start < 0 {
				start = 1
			} else {
				start += 1
			}
		}
	} else {
		if start == 0 {
			start = 1
		} else {
			start = int(math.Abs(float64(start)))
		}
		end = int(math.Abs(api.GetNumber(2)))
		end = int(math.Min(float64(end), float64(size)))
	}

	sub := ""
	for i := start; i <= end; i++ {
		sub += string(cStr[i-1])
	}

	api.PushString(sub)
	return 1
}

func upper(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTString) {
		return 0
	}

	str := api.GetString(0)
	size := str.GetLength()
	cStr := str.GetCStr()

	upper := ""
	for i := 0; i < size; i++ {
		upper += strings.ToUpper(string(cStr[i]))
	}

	api.PushString(upper)
	return 1
}

func RegisterLibString(state *State) {
	lib := NewLibrary(state)
	str := [7]TableMemberReg{
		*NewTableMemberRegCFunction("byte", abyte),
		*NewTableMemberRegCFunction("char", char),
		*NewTableMemberRegCFunction("len", alen),
		*NewTableMemberRegCFunction("lower", lower),
		*NewTableMemberRegCFunction("reverse", reverse),
		*NewTableMemberRegCFunction("sub", sub),
		*NewTableMemberRegCFunction("upper", upper),
	}
	lib.RegisterTableFunction("string", &str[0], len(str))
}
