package io

import (
	. "InterpreterVM/Source/vm"
	"fmt"
	"io"
	"os"
	"unsafe"
)

const MetatableFile = "file"

// For close userdata file
func closeFile(data unsafe.Pointer) {
	err := (*os.File)(data).Close()
	fmt.Println(err)
}

// Helper function for report strerror
func pushError(api *StackAPI, err error) int {
	api.PushNil()
	api.PushString(fmt.Sprint(err))
	return 2
}

// Read by bytes for userdata file
func readBytes(api *StackAPI, file *os.File, bytes int) {
	if bytes <= 0 {
		api.PushString("")
	} else {
		var buf [1024]byte
		n, err := file.Read(buf[:])
		if err != nil {
			fmt.Println(err)
		}
		if n == 0 {
			api.PushNil()
		} else {
			api.PushString(string(buf[:n]))
		}
	}
}

// Read by format for userdata file
func readByFormat(api *StackAPI, file *os.File, format string) {
	// TODO

}

func ioclose(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTUserData) {
		return 0
	}

	userData := api.GetUserData(0)
	closeFile(userData.GetData())
	userData.MarkDestroyed()
	return 0
}

func flush(state *State) int {
	//api := NewStackAPI(state)
	//if !api.CheckArgs(1, ValueTUserData) {
	//	return 0
	//}
	//
	//userData := api.GetUserData(0)
	//(*os.File)(userData.GetData())
	// TODO
	return -1
}

func read(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTUserData) {
		return 0
	}

	userData := api.GetUserData(0)
	file := (*os.File)(userData.GetData())

	params := api.GetStackSize()
	for i := 1; i < params; i++ {
		vType := api.GetValueType(i)
		if vType == ValueTNumber {
			bytes := int(api.GetNumber(i))
			readBytes(api, file, bytes)
		} else if vType == ValueTString {
			format := api.GetString(i).GetStdString()
			readByFormat(api, file, format)
		} else {
			api.PushNil()
		}
	}

	return params - 1
}

func seek(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(1, ValueTUserData, ValueTString, ValueTNumber) {
		return 0
	}

	userData := api.GetUserData(0)
	file := (*os.File)(userData.GetData())

	params := api.GetStackSize()
	if params > 1 {
		whence := api.GetString(1).GetStdString()
		var offset int64 = 0
		if params > 2 {
			offset = int64(api.GetNumber(2))
		}

		switch whence {
		case "set":
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				pushError(api, err)
			}
		case "cur":
			if _, err := file.Seek(offset, io.SeekCurrent); err != nil {
				pushError(api, err)
			}
		case "end":
			if _, err := file.Seek(offset, io.SeekEnd); err != nil {
				pushError(api, err)
			}
		}
	}

	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return pushError(api, err)
	}

	api.PushNumber(float64(pos))
	return 1
}

func setvbuf(state *State) int {
	// TODO
	return -1
}

func write(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTUserData) {
		return 0
	}

	userData := api.GetUserData(0)
	file := (*os.File)(userData.GetData())
	params := api.GetStackSize()
	for i := 1; i < params; i++ {
		vType := api.GetValueType(i)
		if vType == ValueTString {
			str := api.GetString(i)
			cStr := str.GetCStr()
			_, err := file.Write([]byte(fmt.Sprint(cStr)))
			if err != nil {
				return pushError(api, err)
			}
		} else if vType == ValueTNumber {
			_, err := file.Write([]byte(fmt.Sprintf("%.14g", api.GetNumber(i))))
			if err != nil {
				return pushError(api, err)
			}
		} else {
			api.ArgTypeError(i, ValueTString)
			return 0
		}
	}

	api.PushUserData(userData)
	return 1
}

func open(state *State) int {
	//api := NewStackAPI(state)
	//if !api.CheckArgs3(1, ValueTString, ValueTString) {
	//	return 0
	//}
	//
	//fileName := api.GetString(0)
	//mode := "r"
	//
	//if api.GetStackSize() > 1 {
	//	mode = api.GetString(1).GetCStr()
	//}
	//
	//file :=
	// TODO
	return -1
}

func stdin(state *State) int {
	api := NewStackAPI(state)
	userData := state.NewUserData()
	metatable := state.GetMetaTable(MetatableFile)
	userData.Set(unsafe.Pointer(os.Stdin), metatable)
	api.PushUserData(userData)
	return 1
}

func stdout(state *State) int {
	api := NewStackAPI(state)
	userData := state.NewUserData()
	metatable := state.GetMetaTable(MetatableFile)
	userData.Set(unsafe.Pointer(os.Stdout), metatable)
	api.PushUserData(userData)
	return 1
}

func stderr(state *State) int {
	api := NewStackAPI(state)
	userData := state.NewUserData()
	metatable := state.GetMetaTable(MetatableFile)
	userData.Set(unsafe.Pointer(os.Stderr), metatable)
	api.PushUserData(userData)
	return 1
}

func RegisterLibIO(state *State) {
	lib := NewLibrary(state)
	file := [6]TableMemberReg{
		*NewTableMemberRegCFunction("close", ioclose),
		*NewTableMemberRegCFunction("flush", flush),
		*NewTableMemberRegCFunction("read", read),
		*NewTableMemberRegCFunction("seek", seek),
		*NewTableMemberRegCFunction("setvbuf", setvbuf),
		*NewTableMemberRegCFunction("write", write),
	}

	lib.RegisterMetatable(MetatableFile, &file[0], len(file))

	io := [4]TableMemberReg{
		*NewTableMemberRegCFunction("open", open),
		*NewTableMemberRegCFunction("stdin", stdin),
		*NewTableMemberRegCFunction("stdout", stdout),
		*NewTableMemberRegCFunction("stderr", stderr),
	}

	lib.RegisterMetatable("io", &io[0], len(io))
}
