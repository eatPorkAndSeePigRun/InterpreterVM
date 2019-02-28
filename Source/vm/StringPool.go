package vm

type StringPool struct {
	temp    String
	strings map[*String]bool // as set[*String]
}

func NewStringPool() *StringPool {
	return &StringPool{temp: *NewString(""), strings: make(map[*String]bool)}
}

// Get string from pool when string is existed,
// otherwise return nil
func (s *StringPool) GetString(str string) *String {
	s.temp.SetValue(str)
	if _, ok := s.strings[&s.temp]; ok {
		return &s.temp
	} else {
		return nil
	}
}

// Add string to pool
func (s *StringPool) AddString(str *String) {
	s.strings[str] = true
}

// Delete string from pool
func (s *StringPool) DeleteString(str *String) {
	delete(s.strings, str)
}
