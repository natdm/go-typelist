package test

type MyStructType struct {
	data  string
	data2 int
}

type MyInterfaceType interface {
	Read()
	Write()
}

const (
	const1 = iota + 1
	const2
)

type thing []string

// SomeAwesomefunc is awesome
func SomeAwesomefunc(s string, y *string) (string, error) {
	return "", nil
}

var myFunc = func(s string) {

}

type SomeFuncDeclMaybe func(s string)

type ch chan string

type ch2 chan<- string
type ch3 <-chan string

// TakesAChannel is a bit more complex.
func TakesAChannel(in <-chan string, something bool, done chan struct{}) error {
	return nil
}

// MyType has methods
type MyTypeStr struct {
}

// String satisfies stringer
func (t *MyTypeStr) String() string {
	return "Hello"
}

// StringNoPtr has no ptr receiver
func (t MyTypeStr) StringNoPtr() string {
	return "Hello"
}

func TestFunc() {

}
