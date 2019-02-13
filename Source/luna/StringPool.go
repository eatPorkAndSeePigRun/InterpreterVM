package luna

type StringPool struct {
	temp    String
	strings map[*String]int
}

// Get string from pool when string is existed,
// otherwise return nil
func (s StringPool) GetString(str string) *String {
	s.temp.SetValue(str)
	if s.strings[&s.temp] == 1 {
		return &s.temp
	} else {
		return nil
	}
}

// Add string to pool
func (s StringPool) AddString(str *String) {
	s.strings[str] = 1
}

// Delete string from pool
func (s StringPool) DeleteString(str *String) {
	delete(s.strings, str)
}
