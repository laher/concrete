package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
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
	f := &ast.File{
		Name:    ast.NewIdent(pkgName),
		Decls:   []ast.Decl{},
		Imports: []*ast.ImportSpec{},
	}
	addImport(f, "flag", "_")
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

func TestNoImplementationExists(t *testing.T) {
	implName := "MyImpl"
	pkgName := "main"
	f := &ast.File{
		Name:    ast.NewIdent(pkgName),
		Decls:   []ast.Decl{},
		Imports: []*ast.ImportSpec{},
	}
	addImplementationStruct(f, implName)
	buf := &bytes.Buffer{}
	err := writeAST(f, buf)
	if err != nil {
		t.Fatalf("Failed ... %s", err)
	}
	out := buf.String()
	if out != `package main

type MyImpl struct {
}
` {
		t.Fatalf("incorrect code generated: |%s|", out)
	}
}

func TestStructPointerImplementationExists(t *testing.T) {
	implName := "MyImpl"
	fset := token.NewFileSet()
	minimal := `package main

type MyImpl struct {
}
`
	f, err := parser.ParseFile(fset, "blah.go", minimal, 0)
	if err != nil {
		t.Fatalf("failed to parse implementation")
	}
	addImplementationStruct(f, implName)
	buf := &bytes.Buffer{}
	err = writeAST(f, buf)
	if err != nil {
		t.Fatalf("Failed ... %s", err)
	}
	out := buf.String()
	expected := `package main

type MyImpl struct{}
`
	if out != expected {
		t.Fatalf("incorrect code generated: |%s| vs |%s|", out, expected)
	}
}
