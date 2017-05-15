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

type FuncTyp func(s string)

type ch chan string
type ch2 chan<- string
type ch3 <-chan string

// TakesAChan is a bit more complex.
func TakesAChan(in <-chan string, something bool, done chan struct{}) error {
	return nil
}
