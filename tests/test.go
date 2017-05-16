package test

// BasicType should be a GenDecl
type BasicType string

//ArrayType should be a GenDecl
type ArrayType [2]string

// SliceType should be a GenDecl
type SliceType []string

// FuncType should be a GenDecl
type FuncType func(foo int, bar string) error

// InterfaceType should be a GenDecl
type InterfaceType interface{}

// StructType should be a GenDecl
type StructType struct{}

var Variable = "TestVar"

// ConstType is currently a should be a GenDecl but should be saved
// as something else since it's a differnet type in syntax (name is red)
const ConstType = 1

// Standalonefunc is not being parsed.. odd.
func Standalonefunc(foo int, bar string) error {
	return nil
}

// MethodDeclPR should show up as a MethodDecl
func (s *StructType) MethodDeclPR(foo int, bar string) error {
	return nil
}

// MethodDecl should show up as a MethodDecl
func (s StructType) MethodDecl(foo int, bar string) error {
	return nil
}

// init is not being parsed..
func init() {
	// init func
}

// init is not being parsed..
func init() {
	// main func
}

func somefunc() {

}
