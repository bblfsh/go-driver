package fixtures

type Foo struct{}

var (
	_ Interf = Foo{}
	_ Interf = &Foo{}
	_ Interf = (*Foo)(nil)
)
