package guintf

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"reflect"
	"strings"
)

func parseFunctions(filename string, Type reflect.Type) []Func {
	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	var m []Func

	for _, f := range node.Decls {
		d, ok := f.(*ast.GenDecl)
		if !ok || d.Tok != token.CONST {
			continue
		}
		fmt.Println(strings.TrimSpace(d.Doc.Text()))
		for _, f := range d.Specs {
			v := f.(*ast.ValueSpec)
			m = append(m, Func{Name: strings.TrimPrefix(v.Names[0].Name, "Msg")})
		}
	}
MethodLoop:
	for j := range m {
		for i := 0; i < Type.NumMethod(); i++ {
			met := Type.Method(i)
			if m[j].Name == met.Name {
				m[j].ParamType = met.Type.In(1)
				continue MethodLoop
			}
		}
		log.Fatal(m[j], ": not found")
	}

	//pkg, err := importer.Default().Import("time")

	return m
}
