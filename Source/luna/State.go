package luna

type State struct {
}

func (state State) GetString(str string) *String {
	return &String{}
}
