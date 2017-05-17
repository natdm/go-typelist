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
		o := inspectNode(file.Decls[i], bs, fset)
		switch o.Type {
		case "":
			log.Println("Shits got no type")
			continue
		default:
			out[o.Line] = &o
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
func inspectNode(node ast.Node, bs []byte, fset *token.FileSet) Object {
	obj := Object{}
	ast.Inspect(node, func(n ast.Node) bool {
		obj.Line = fset.File(node.Pos()).Line(node.Pos())
		switch x := n.(type) {
		case *ast.FuncDecl:
			body := getbody(bs, node)
			index := strings.Index(body, "{\n")
			var sig string
			var name string
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
		case *ast.ChanType:
			obj.Signature = getSignature(bs, node)
			obj.Type = "ChanType"
		case *ast.DeclStmt:
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
		case *ast.FuncLit:
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
		case *ast.FuncType:
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
		case *ast.StructType:
			body := getbody(bs, node)
			index := strings.Index(body, "struct")
			obj.Signature = strings.TrimSuffix(body[:index+6], "\n")
			obj.Type = "StructType"
		case *ast.TypeSpec:
			obj.Signature = getSignature(bs, node)
			obj.Type = "TypeSpec"
		case *ast.GenDecl:
			if strings.Contains(string(bs[node.Pos()-1:node.End()]), "import") {
				break
			}
			body := getbody(bs, node)
			idx := strings.Index(body, "{\n")
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
			obj.Signature = sig
			obj.Type = "GenDecl"

			idx = strings.Index(sig, " ")
			log.Println(idx)
			pfx := sig[:idx]
			log.Println(pfx)
			switch pfx {
			case "type", "var", "const":
				start := strings.TrimSpace(sig[idx:])
				log.Printf("%s start\n", start)
				endIdx := strings.Index(start, " ")
				if endIdx == -1 { // no idea how
					log.Printf("%s has no ending space\n", start)
					return false
				}
				name := start[:endIdx]
				obj.Name = &name
			}
		default:
			// log.Println(string(bs[x.Pos()-1 : x.End()]))
		}
		return false
	})
	return obj
}
