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

	"golang.org/x/text/unicode/norm"
)

// Object represents a Go type
type Object struct {
	Line      int    `json:"line"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Signature string `json:"signature"`
	Type      string `json:"type"`
}

func main() {
	if len(os.Args) < 1 {
		log.Println("Nothing to do")
		os.Exit(0)
	}
	out := []Object{}
	fset := token.NewFileSet()
	file, e := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if e != nil {
		log.Fatal(e)
		return
	}
	bs, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(e)
	}

	for k, v := range file.Scope.Objects {
		switch v.Kind {
		case ast.Typ:
			log.Printf("%s is an ast type\n", k)
		case ast.Fun:
			log.Printf("%s is a func type\n", k)
		case ast.Var:
			log.Printf("%s is a var type\n", k)
		default:
			log.Printf("%s is not a known type\n", k)
		}
		ts, ok := v.Decl.(*ast.TypeSpec)
		if ok {
			obj := Object{
				Name: k,
				Line: fset.File(v.Pos()).Line(v.Pos()),
			}

			Type(bs, ts, &obj)
			out = append(out, obj)
		}
		fd, ok := v.Decl.(*ast.FuncDecl)
		if ok {
			out = append(out, Object{
				Type:      "FuncDecl",
				Name:      k,
				Signature: string(bs[fd.Type.Pos()-1 : fd.Type.End()]),
				Line:      fset.File(v.Pos()).Line(v.Pos()),
				Comment:   "I am a comment",
			})
		}
		fl, ok := v.Decl.(*ast.FuncLit)
		if ok {
			out = append(out, Object{
				Type:      "FuncLit",
				Name:      k,
				Signature: string(bs[fl.Pos()-1 : fl.End()]),
				Line:      fset.File(v.Pos()).Line(v.Pos()),
				Comment:   "I am a comment",
			})
		}
		ft, ok := v.Decl.(*ast.FuncType)
		if ok {
			out = append(out, Object{
				Type:      "FuncType",
				Name:      k,
				Signature: string(bs[ft.Pos()-1 : ft.End()]),
				Line:      fset.File(v.Pos()).Line(v.Pos()),
				Comment:   "I am a comment",
			})
		}
	}
	outBs, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, string(norm.NFC.Bytes(outBs)))
}

// Type creates a package level type.
func Type(bs []byte, ts *ast.TypeSpec, obj *Object) error {
	var comment string
	if ts.Comment != nil {
		comment = ts.Comment.Text()
	}
	obj.Comment = comment

	switch ts.Type.(type) {
	case *ast.ChanType:
		x := ts.Type.(*ast.ChanType)
		obj.Type = "ChanType"
		obj.Signature = string(bs[x.Pos()-1 : x.End()])
	case *ast.InterfaceType:
		x := ts.Type.(*ast.InterfaceType)
		obj.Type = "InterfaceType"
		obj.Signature = string(bs[x.Pos()-1 : x.End()])
	case *ast.ArrayType:
		x := ts.Type.(*ast.ArrayType)
		obj.Type = "ArrayType"
		obj.Signature = string(bs[x.Pos()-1 : x.End()])
	case *ast.MapType:
		x := ts.Type.(*ast.MapType)
		obj.Type = "MapType"
		obj.Signature = string(bs[x.Pos()-1 : x.End()])
	case *ast.StructType:
		x := ts.Type.(*ast.StructType)
		obj.Type = "StructType"
		obj.Signature = string(bs[x.Pos()-1 : x.End()])
	default:
		obj.Type = "Unknown"
	}
	return nil
}
