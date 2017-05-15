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
	Line      int    `json:"line"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Signature string `json:"signature"`
	Type      string `json:"type"`
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

	out := []Object{}
	fset := token.NewFileSet()
	file, e := parser.ParseFile(fset, os.Args[1], nil, parser.AllErrors)
	if e != nil {
		log.Fatal(e)
		return
	}
	bs, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(e)
	}

	for i := range file.Decls {
		o := inspectNode(file.Decls[i], bs, fset)
		if o.Type == "MethodDecl" {
			out = append(out, o)
		}

	}

	for k, v := range file.Scope.Objects {
		fld, ok := v.Decl.(*ast.Field)
		if ok {
			obj := Object{
				Name:      k,
				Signature: string(bs[fld.Pos()-1 : fld.End()]),
				Line:      fset.File(v.Pos()).Line(v.Pos()),
			}
			out = append(out, obj)
			continue
		}

		se, ok := v.Decl.(*ast.SelectorExpr)
		if ok {
			obj := Object{
				Name:      k,
				Signature: string(bs[se.Pos()-1 : se.End()]),
				Line:      fset.File(v.Pos()).Line(v.Pos()),
			}
			out = append(out, obj)
			continue
		}

		ts, ok := v.Decl.(*ast.TypeSpec)
		if ok {
			obj := Object{
				Name:      k,
				Line:      fset.File(v.Pos()).Line(v.Pos()),
				Signature: string(bs[ts.Pos()-1 : ts.End()]),
				Type:      "TypeSpec",
			}
			out = append(out, obj)
			continue
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
			continue
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
			continue
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
			continue
		}

	}
	objV := ObjectsVersion{Version: version, Objects: out}
	outBs, err := json.MarshalIndent(objV, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, string(norm.NFC.Bytes(outBs)))
}

// inspectNode checks what is determined to be the value of a node based on a type-assertion.
func inspectNode(node ast.Node, bs []byte, fset *token.FileSet) Object {
	obj := Object{}
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Recv == nil {
				return true
			}
			body := string(bs[x.Pos()-1 : x.End()])
			index := strings.Index(body, "{\n\treturn")
			sig := body[:index-1]
			obj.Name = x.Name.String()
			obj.Signature = sig
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "MethodDecl"
		case *ast.ChanType:
			obj.Name = ""
			obj.Signature = string(bs[x.Pos()-1 : x.End()])
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "ChanType"
		default:
		}
		return true
	})
	return obj
}
