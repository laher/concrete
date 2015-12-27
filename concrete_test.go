package main

import (
	"bytes"
	"go/ast"
	"os"
	"testing"
)

func TestConcreteSimple(t *testing.T) {
	const input = `
package temperature
import "fmt"
type Celsius float64
func (c Celsius) String() string  { return fmt.Sprintf("%g°C", c) }
func (c *Celsius) SetF(f float64) { *c = Celsius(f - 32 / 9 * 5) }

type Rdr interface {
	Get() (string, error)
	Set(string) error
}
	`

	interfaceName := "Rdr"
	concreteType := "My" + interfaceName
	pkgName := "temperature"
	parseAndPrint(os.Stdout, input, interfaceName, concreteType, pkgName)
}

func TestAst1(t *testing.T) {
	implName := "MyImpl"
	pkgName := "main"
	/*
		f, err := cr(os.Stdout, implName, "mypkg")
		if err != nil {
			return err
		}
	*/
	f := &ast.File{
		Name:    ast.NewIdent(pkgName),
		Decls:   []ast.Decl{},
		Imports: []*ast.ImportSpec{},
	}

	/*
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "blah.go", minimal, 0)
		if err != nil {
			return err
		}
	*/
	addImport(f, "flag", "_")

	//	fmt.Printf("File: %+v", f)
	/*
		for _, d := range f.Decls {
			switch f := d.(type) {
			case *ast.GenDecl:
				for _, s := range f.Specs {
					switch t := s.(type) {
					case *ast.TypeSpec:
						//ok. found a type. ?
						fmt.Printf("type name: %s\n", t.Name)
					}
				}
			case *ast.FuncDecl:
			}
		}
	*/
	addImplementationStruct(f, implName)
	buf := &bytes.Buffer{}
	err := writeAST(f, buf)
	if err != nil {
		t.Fatalf("Failed ... %s", err)
	}
	out := buf.String()
	if out != `package main

import _ "flag"

type MyImpl struct {
}
` {
		t.Fatalf("incorrect code generated: |%s|", out)
	}
}
