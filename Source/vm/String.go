package vm

type String struct {
	gcObjectField
	inHeap    byte   // String in heap or not
	strBuffer string // Buffer for short string
	str       string // Pointer to heap which stored long string
	length    int    // Length of string
	hash_     int64  // Hash value of string
}

func NewString(str string) *String {
	var s String
	s.SetValue(str)
	return &s
}

// Calculate hash of string
func (s *String) hash(str string) {
	s.hash_ = 5381

	for _, c := range str {
		s.hash_ = ((s.hash_ << 5) + s.hash_) + int64(c)
	}
}

func (s *String) Accept(v GCObjectVisitor) {
	v.VisitString(s)
}

func (s *String) GetHash() int64 {
	return s.hash_
}

func (s *String) GetLength() int {
	return s.length
}

func (s *String) GetCStr() string {
	if s.inHeap != 0 {
		return s.str
	} else {
		return s.strBuffer
	}
}

// Convert to string
func (s *String) GetStdString() string {
	if s.inHeap != 0 {
		return s.str[0:s.length]
	} else {
		return s.strBuffer[0:s.length]
	}
}

// Change context of string
func (s *String) SetValue(str string) {
	if s.inHeap != 0 {
		s.str = ""
	}
	s.length = len(str)
	if s.length < len(s.strBuffer) {
		s.strBuffer = s.str
		s.inHeap = 0
		s.hash(s.strBuffer)
	} else {
		s.str = str
		s.inHeap = 1
		s.hash(s.str)
	}
}

func (s *String) IsEqual(s1 String) bool {
	if s.inHeap != 0 && s.str != s1.str {
		return false
	} else if s.inHeap != 0 && s.strBuffer != s1.strBuffer {
		return false
	}
	return s.hash_ == s1.hash_ && s.length == s1.length
}

func (s *String) IsLess(s1 String) bool {
	var s_, s1_ string
	if s.inHeap != 0 {
		s_ = s.str
	} else {
		s_ = s.strBuffer
	}
	if s1.inHeap != 0 {
		s1_ = s1.str
	} else {
		s1_ = s1.strBuffer
	}
	if s_ == s1_ {
		return s.length < s1.length
	} else {
		return s1_ < s_
	}
}
