package primitives

type Label *string

func NewLabel(str string) Label {
	return func(str string) *string { return &str }(str)
}
