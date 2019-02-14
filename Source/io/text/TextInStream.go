package text

import (
	"os"
)

type InStream struct {
	stream *os.File
}

func (is InStream) IsOpen() bool {
	return is.stream != nil
}

func (is InStream) GetChar() uint8 {
	buf := make([]byte, 1)
	_, err := is.stream.Read(buf)
	panic(err)
	return uint8(buf[0])
}

type InStringStream struct {
	str string
	pos int
}

func NewInStringStream(str string) InStringStream {
	return InStringStream{str, 0}
}

func (iss InStringStream) SetInputString(input string) {
	iss.str = input
	iss.pos = 0
}

func (iss InStringStream) GetChar() uint8 {
	if iss.pos < len(iss.str) {
		res := iss.str[iss.pos]
		iss.pos++
		return res
	} else {
		return 0
	}
}
