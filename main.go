package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Object represents a Go type
type Object struct {
	Line      int       `json:"line"`
	Signature string    `json:"signature"`
	Type      string    `json:"type"`
	Name      *string   `json:"name"`
	Receiver  *Receiver `json:"receiver"`
}

type Receiver struct {
	TypeName string `json:"type_name"`
	Pointer  bool   `json:"pointer"`
	Alias    string `json:"alias"`
}

func (r *Receiver) String() string {
	if r.Pointer {
		return fmt.Sprintf("(%s *%s)", r.Alias, r.TypeName)
	}
	return fmt.Sprintf("(%s %s)", r.Alias, r.TypeName)
}

// ObjectsVersion has the version of the package as well as the objects
type ObjectsVersion struct {
	Version string   `json:"version"`
	Objects []Object `json:"objects"`
}

const version = "0.0.1"

func main() {
	if len(os.Args) < 1 {
		log.Println("Nothing to do")
		os.Exit(0)
	}

	// map the line number to the object
	out := map[int]*Object{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, os.Args[1], nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	bs, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// get all signatures, types, and lines and add them to the map
	for i := range file.Decls {
		for _, o := range inspectNode(file.Decls[i], bs, fset) {
			switch o.Type {
			case "":
				log.Println("no type")
				continue
			default:
				out[o.Line] = &o
			}
		}
	}

	// collect the names of these types for syntax highlighting.
	for k, v := range file.Scope.Objects {
		loc := fset.File(v.Pos()).Line(v.Pos())
		if _, ok := out[loc]; ok {
			switch out[loc].Type {
			case "StructType", "FuncType":
				n := k
				out[loc].Name = &n
			}
		}
	}

	// add names of methods to Name
	for k := range out {
		if out[k].Receiver != nil && out[k].Name == nil {
			s := strings.TrimPrefix(out[k].Signature, "func "+out[k].Receiver.String())
			paren := strings.Index(s, "(")
			x := strings.TrimSpace(s[:paren])
			out[k].Name = &x
		}
	}

	objArr := []Object{}
	for _, v := range out {
		objArr = append(objArr, *v)
	}
	objV := ObjectsVersion{Version: version, Objects: objArr}
	outBs, err := json.MarshalIndent(objV, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, string(norm.NFC.Bytes(outBs)))
}

func getbody(bs []byte, node ast.Node) string {
	return string(bs[node.Pos()-1 : node.End()])
}

func getSignature(bs []byte, node ast.Node) string {
	return strings.TrimSuffix(getbody(bs, node), "\n")
}

func parseReceiver(r string) *Receiver {
	ptr := strings.Index(r, "*")
	a := strings.TrimPrefix(r, "(")
	aSep := strings.Index(a, " ")
	ptrB := false
	var typName string
	if ptr == -1 { // not a pointer
		typName = strings.TrimSuffix(a[aSep+1:], ")")
	} else {
		typName = strings.TrimSuffix(a[aSep+2:], ")")
		ptrB = true
	}
	return &Receiver{
		Pointer:  ptrB,
		TypeName: typName,
		Alias:    a[:aSep],
	}
}

// inspectNode checks what is determined to be the value of a node based on a type-assertion.
// returns multiple in the case of grouped types (const, var, type, etc)
func inspectNode(node ast.Node, bs []byte, fset *token.FileSet) []Object {
	out := []Object{}
	ast.Inspect(node, func(n ast.Node) bool {
		line := fset.File(node.Pos()).Line(node.Pos())
		switch x := n.(type) {
		case *ast.FuncDecl:
			body := getbody(bs, node)
			index := strings.Index(body, "{\n")
			var sig string
			var name string
			var obj Object
			if index > -1 {
				sig = body[:index-1]
			} else {
				sig = body
			}
			obj.Signature = sig
			if x.Recv == nil {
				obj.Type = "FuncDecl"
				name = strings.TrimPrefix(sig, "func")
			} else {
				obj.Type = "MethodDecl"
				rcv := string(bs[x.Recv.Opening-1 : x.Recv.Closing])
				obj.Receiver = parseReceiver(rcv)
				name = strings.TrimPrefix(sig, "func"+" "+rcv)
			}
			index = strings.Index(name, "(")
			name = strings.TrimSpace(name[:index])
			obj.Name = &name
			out = append(out, obj)
		case *ast.ChanType:
			var obj Object
			obj.Signature = getSignature(bs, node)
			obj.Type = "ChanType"
			out = append(out, obj)
		case *ast.DeclStmt:
			var obj Object
			body := getbody(bs, node)
			index := strings.Index(body, "{\n")
			var sig string
			if index > -1 {
				sig = body[:index-1]
			} else {
				sig = body
			}
			obj.Signature = strings.TrimSuffix(sig, "\n")
			obj.Type = "DeclStmt"
			out = append(out, obj)
		case *ast.FuncLit:
			var obj Object
			body := getbody(bs, node)
			index := strings.Index(body, "{\n")
			var sig string
			if index > -1 {
				sig = body[:index-1]
			} else {
				sig = body
			}
			obj.Signature = strings.TrimSuffix(sig, "\n")
			obj.Type = "FuncLit"
			out = append(out, obj)
		case *ast.FuncType:
			var obj Object
			body := getbody(bs, node)
			index := strings.Index(body, "{\n")
			var sig string
			if index > -1 {
				sig = body[:index-1]
			} else {
				sig = body
			}
			obj.Signature = strings.TrimSuffix(sig, "\n")
			obj.Type = "FuncType"
			out = append(out, obj)
		case *ast.StructType:
			var obj Object
			body := getbody(bs, node)
			index := strings.Index(body, "struct")
			obj.Signature = strings.TrimSuffix(body[:index+6], "\n")
			obj.Type = "StructType"
			out = append(out, obj)
		case *ast.TypeSpec:
			var obj Object
			obj.Signature = strings.TrimSpace(getSignature(bs, node))
			obj.Type = "TypeSpec"
			out = append(out, obj)
		case *ast.GenDecl:
			if strings.Contains(string(bs[node.Pos()-1:node.End()]), "import") {
				break
			}

			body := getbody(bs, node)
			idx := strings.Index(body, "=")
			// if no equal in "var x = ___" check for other endings
			if idx == -1 {
				idx = strings.Index(body, "{\n")
			}

			if idx == -1 {
				// special case for inline structs
				idx = strings.Index(body, "{")
			}
			var sig string
			if idx > -1 {
				sig = body[:idx]
			} else {
				sig = body
			}

			idx = strings.Index(sig, " ")
			pfx := sig[:idx]
			switch pfx {
			case "type", "var", "const":
				start := strings.TrimSpace(sig[idx:])
				endIdx := strings.Index(start, " ")
				if endIdx == -1 {
					// happens when grouped
					base := string(bs[x.Pos()-1 : x.End()])
					split := strings.Split(base, "\n")
					for i, v := range split {
						if i == 0 || i > len(split)-3 {
							// skip the first 'const (' line followed by the )
							continue
						}
						if v == "" {
							// skip any gaps
							continue
						}
						newObj := Object{}
						_name := strings.TrimSpace(v)
						if eqidx := strings.Index(_name, "="); eqidx > -1 {
							_name = strings.TrimSpace(_name[:eqidx])
						}
						newObj.Signature = fmt.Sprintf("%s %s", pfx, _name)
						newObj.Name = &_name
						newObj.Line = line + i
						newObj.Type = "GenDecl"
						log.Printf("%+v\n", newObj)
						out = append(out, newObj)
					}
					return true
				}
				var obj Object
				obj.Signature = strings.TrimSpace(sig)
				obj.Type = "GenDecl"
				name := strings.TrimSpace(start[:endIdx])
				obj.Name = &name
				obj.Line = line
				out = append(out, obj)

			}
		default:
			// log.Println(string(bs[x.Pos()-1 : x.End()]))
		}
		return false
	})
	return out
}
