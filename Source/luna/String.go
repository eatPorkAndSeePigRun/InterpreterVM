package luna

type String struct {
	inHeap    string  // Calculate hash of string
	strBuffer string  // Buffer for short string
	str       *string // Pointer to heap which stored long string
	length    uint64  // Length of string
	hash      int64   // Hash value of string
}

func (str String) hash_(s string)  {
	
}

func (str String) GetCStr() string {
	return ""
}

func (str String) GetStdString() string {
	return ""
}
