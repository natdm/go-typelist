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
	Line      int     `json:"line"`
	Signature string  `json:"signature"`
	Type      string  `json:"type"`
	Name      *string `json:"name"`
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
			continue
		default:
			out[o.Line] = &o
		}
	}

	// collect the names of these types for syntax highlighting.
	for _, v := range file.Scope.Objects {
		switch v.Decl.(type) {
		case *ast.Field,
			*ast.SelectorExpr,
			*ast.TypeSpec,
			*ast.FuncDecl,
			*ast.FuncLit,
			*ast.FuncType:
			out[fset.File(v.Pos()).Line(v.Pos())].Name = &v.Name
		default:
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
			obj.Signature = sig
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "MethodDecl"
		case *ast.ChanType:
			obj.Signature = string(bs[node.Pos()-1 : node.End()])
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "ChanType"
		case *ast.DeclStmt:
			body := string(bs[x.Pos()-1 : x.End()])
			index := strings.Index(body, "{\n")
			sig := body[:index-1]
			z := n.(*ast.DeclStmt)
			obj.Signature = sig
			obj.Line = fset.File(z.Pos()).Line(z.Pos())
			obj.Type = "DeclStmt"
		case *ast.FuncLit:
			body := string(bs[node.Pos()-1 : node.End()])
			index := strings.Index(body, "{\n")
			sig := body[:index-1]
			z := n.(*ast.FuncLit)
			obj.Signature = sig
			obj.Line = fset.File(z.Type.Pos()).Line(z.Type.Pos())
			obj.Type = "FuncLit"
		case *ast.FuncType:
			body := string(bs[node.Pos()-1 : node.End()])
			index := strings.Index(body, "{")
			var sig string
			if index > -1 {
				sig = body[:index-1]
			} else {
				sig = body
			}
			obj.Signature = sig
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "FuncType"
		case *ast.StructType:
			body := string(bs[node.Pos()-1 : node.End()])
			index := strings.Index(body, "struct")
			sig := body[:index+6]
			obj.Signature = sig
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "StructType"
		case *ast.TypeSpec:
			obj.Signature = string(bs[node.Pos()-1 : node.End()])
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "TypeSpec"
		case *ast.GenDecl:
			obj.Signature = string(bs[node.Pos()-1 : node.End()])
			obj.Line = fset.File(x.Pos()).Line(x.Pos())
			obj.Type = "TypeSpec"
		default:
		}
		return true
	})
	return obj
}
