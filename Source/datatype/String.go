package datatype

type String struct {
	GCObject
	inHeap    uint8  // String in heap or not
	strBuffer string // Buffer for short string
	str       string // Pointer to heap which stored long string
	length    int    // Length of string
	hash_     int64  // Hash value of string
}

func NewString(str string) String {
	var s String
	s.SetValue(str)
	return s
}

// Calculate hash of string
func (s *String) hash(str string) {
	s.hash_ = 5381

	for _, c := range str {
		s.hash_ = ((s.hash_ << 5) + s.hash_) + int64(c)
	}
}

func (s String) Accept(v GCObjectVisitor) {
	v.VisitString(&s)
}

func (s String) GetHash() int64 {
	return s.hash_
}

func (s String) GetLength() int {
	return s.length
}

func (s String) GetCStr() string {
	if s.inHeap != 0 {
		return s.str
	} else {
		return s.strBuffer
	}
}

// Convert to string
func (s String) GetStdString() string {
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
