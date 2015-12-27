package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"text/template"
)

func doMain2() error {
	implName := "MyImpl"
	/*
			f, err := cr(os.Stdout, implName, "mypkg")
			if err != nil {
				return err
			}
		f := &ast.File{
			Name:    ast.NewIdent("mypkg"),
			Decls:   []ast.Decl{},
			Imports: []*ast.ImportSpec{},
		}
	*/
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "blah.go", minimal, 0)
	if err != nil {
		return err
	}
	i := &ast.ImportSpec{Path: &ast.BasicLit{Value: "\"flag\""}, Name: ast.NewIdent("_")}
	i.Name.NamePos = token.Pos(2)
	f.Imports = append(f.Imports, i)
	//	fmt.Printf("File: %+v", f)
	for _, i := range f.Imports {
		fmt.Printf("import: %#v\n", i)
		fmt.Printf("import: %s - %s\n", i.Name, i.Path.Value)
		if i.Name != nil {
			fmt.Printf("import Name: %#v \n", i.Name)
		}
	}
	fmt.Printf("package: %#v\n", f.Package)
	f.Name.Name = "m"
	//	f.Package = "m"
	for _, d := range f.Decls {
		fmt.Printf("decl: %#v\n", d)
		fmt.Printf("\n")
	}
	implFound := false
	ast.Inspect(f, func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.TypeSpec:
			//ok. found a type. ?
			if t.Name.Name == implName {
				implFound = true
			}
		}
		return true
	})
	if !implFound {
		//generate struct
		n := ast.NewIdent(implName)
		n.Obj = &ast.Object{Name: implName, Kind: ast.Typ}
		ts := &ast.TypeSpec{Name: n, Type: &ast.StructType{
			Fields: &ast.FieldList{},
		}}
		n.Obj.Decl = ts
		gd := &ast.GenDecl{Specs: []ast.Spec{
			ts,
		}, Tok: token.TYPE}
		f.Decls = append(f.Decls, gd)

	} else {
		//already found. No generaty
	}
	fmt.Printf("\n\n")
	ast.Inspect(f, func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.TypeSpec:
			//ok. found a type. ?
			fmt.Printf("type (%p): %#v\n", t, t)
			fmt.Printf("type.Name: %#v\n", t.Name)
			fmt.Printf("type.Name.Obj: %#v\n", t.Name.Obj)
			fmt.Printf("type.Name.Obj.Decl: %#v\n", t.Name.Obj.Decl)

			fmt.Printf("type.Type: %#v\n", t.Type)
			switch st := t.Type.(type) {
			case *ast.StructType:
				fmt.Printf("type.Type.Fields: %#v\n", st.Fields)
			}
			fmt.Println("")
		}
		return true
	})

	for _, d := range f.Decls {
		fmt.Printf("decl (%p): %#v\n", d, d)
	}
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
	return write("myimpl.gox", f)
}

func cr(w io.Writer, implementation, pkg string) (*ast.File, error) {
	input, err := createBasic(w, implementation, pkg)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "blah.go", input, 0)
	if err != nil {
		return nil, err
	}

	return f, nil
}

const minimal = `package x

`
const minimal2 = `package x

type Z struct{}
`

var implementationTemplate = `package {{.Pkg}}

	type {{.Implementation}} struct {}
`
var implementationTemplate2 = `package {{.Pkg}}

import(
	"fmt"
	"io"
	_ "flag"
)

type {{.Implementation}} struct {}

func (i {{.Implementation}}) Do(w io.Writer) error {
	return fmt.Errorf("Not implemented")
}
`

func createBasic(w io.Writer, implementation, pkg string) (string, error) {

	tmpl, err := template.New("implementationTemplate").Parse(implementationTemplate2)
	if err != nil {
		fmt.Fprintf(w, "Error parsing template")
		return "", err
	}
	data := struct {
		Pkg            string
		Implementation string
	}{
		Implementation: implementation,
		Pkg:            pkg,
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		fmt.Fprintf(w, "Error processing template")
		return "", err
	}
	typeDecl := out.String()
	return typeDecl, nil
}

func write(filename string, f *ast.File) error {
	fset := token.NewFileSet()
	fw, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fw.Close()
	err = (&printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}).Fprint(fw, fset, f)
	if err != nil {
		return err
	}
	err = fw.Close()
	if err != nil {
		return err
	}
	return nil
}
