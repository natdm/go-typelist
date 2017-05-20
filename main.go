package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

// Object represents a Go type
type Object struct {
	Line      int       `json:"line"`
	Signature string    `json:"signature"`
	Type      string    `json:"type"`
	Name      *string   `json:"name"`
	Receiver  *Receiver `json:"receiver"`
}

// Receiver is a pointer receiver
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

// Sort ...
func (o *ObjectsVersion) Sort() {
	sort.Slice(o.Objects, func(i, j int) bool {
		return o.Objects[i].Line < o.Objects[j].Line
	})
}

const version = "0.0.2"

func main() {
	flagVersion := flag.Bool("v", false, "Print version")
	flag.Usage = usage
	flag.Parse()

	if *flagVersion {
		fmt.Fprintln(os.Stdout, version)
		os.Exit(0)
	}

	if len(os.Args) < 2 || !strings.HasSuffix(os.Args[1], ".go") {
		log.Fatalln("Need a Go file")
	}

	out, err := parse(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprint(os.Stdout, out)
}

func parse(f string) (string, error) {
	// map the line number to the object
	out := map[int]*Object{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f, nil, parser.AllErrors)
	if err != nil {
		return "", err
	}
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}

	// get all signatures, types, and lines and add them to the map
	for i := range file.Decls {
		objs := inspectNode(file.Decls[i], bs, fset)
		for i := range objs {
			switch objs[i].Type {
			case "":
				log.Println("no type")
				continue
			default:
				out[objs[i].Line] = &objs[i]
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
	objV.Sort()
	outBs, err := json.MarshalIndent(objV, "", "    ")
	if err != nil {
		return "", err
	}
	return string(outBs), nil
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
			// further trim name down
			name = strings.TrimSpace(name[:index])
			obj.Name = &name
			obj.Line = line
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

			brIdx := strings.Index(body, "\n")
			eqIdx := strings.Index(body, "=")
			if eqIdx == -1 {
				// is a type alias

				spaceSplit := strings.Split(body, " ")

				var sig string
				if strings.Contains(spaceSplit[2], "func") {
					sig = strings.TrimSuffix(body, "\n")
				} else {
					sig = fmt.Sprintf("%s %s %s", spaceSplit[0], spaceSplit[1], strings.TrimSuffix(spaceSplit[2], "\n"))
				}

				// signature =
				out = append(out, Object{
					Name:      &spaceSplit[1],
					Type:      "GenDecl",
					Line:      line,
					Signature: sig,
				})
				return false
			} else if brIdx > eqIdx {
				// is a var x = __
				spaceSplit := strings.Split(body[:eqIdx], " ")

				// signature =
				out = append(out, Object{
					Name:      &spaceSplit[1],
					Type:      "GenDecl",
					Line:      line,
					Signature: fmt.Sprintf("%s %s", spaceSplit[0], spaceSplit[1]),
				})
				return false
			} else {
				// is a "const ()" entry
				// find if var, or const, or type
				declType := strings.Split(body, " ")[0]
				split := strings.Split(body, "\n")
				for i, v := range split {
					if i == 0 || v == "" || v == "(" || v == ")" {
						continue
					}
					if strings.TrimSpace(strings.Split(v, " ")[0]) == "//" {
						continue
					}
					name := strings.TrimSpace(strings.Split(v, "=")[0])
					if strings.HasPrefix(name, "//") {
						continue
					}
					out = append(out, Object{
						Line:      line + i,
						Name:      &name,
						Type:      "GenDecl",
						Signature: fmt.Sprintf("%s %s", declType, name),
					})
				}
				return false
			}
		default:
			// log.Println(string(bs[x.Pos()-1 : x.End()]))
		}
		return false
	})
	return out
}
