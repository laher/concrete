package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	interfaceName   = flag.String("interface", "", "The name of the interface to be implemented")
	inPackage       = flag.String("in-package", ".", "package where interface is defined")
	implPackage     = flag.String("impl-package", ".", "package where implementation should be created")
	concreteTypeTpl = flag.String("implementation", "{{.Interface}}Impl", "The name of the concrete implementation")
	helpFlag        = flag.Bool("help", false, "show detailed help message")
	writeFlag       = flag.Bool("w", false, "rewrite input files in place (by default, the results are printed to standard output)")
	verboseFlag     = flag.Bool("v", false, "show verbose matcher diagnostics")
	listInterfaces  = flag.Bool("l", false, "list interfaces")
)

const usage = `concrete: a tool to implement Go interfaces

Usage: concrete -interface <InterfaceName> [options]

`

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "concrete: %s\n", err)
		os.Exit(1)
	}
}

func doMain() error {
	flag.Parse()
	w := os.Stdout
	if *listInterfaces {
		inPkg, err := packageNameToPkg(*inPackage)
		if err != nil {
			return err
		}
		names := inPkg.Scope().Names()
		for _, name := range names {
			lu := inPkg.Scope().Lookup(name)
			nu := lu.Type().Underlying()

			_, ok := nu.(*types.Interface)
			if ok {
				fmt.Fprintf(w, " %s\n", lu.Name())
			}
		}
		return nil
	}
	if *helpFlag || *interfaceName == "" {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		os.Exit(2)
	}

	tmpl, err := template.New("concreteTypeTemplate").Parse(*concreteTypeTpl)
	if err != nil {
		fmt.Fprintf(w, "Error parsing template")
		return err
	}
	data := struct {
		Interface string
	}{
		Interface: *interfaceName,
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		fmt.Fprintf(w, "Error processing template")
		return err
	}
	concreteType := out.String()
	if *writeFlag {
		w, err = os.OpenFile(concreteType+".gox", os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return err
		}
	}
	return parseAndPrintFiles(w, *inPackage, *interfaceName, *implPackage, out.String())
}

func pkgToFiles(pkg string) (*token.FileSet, []*ast.File, error) {
	if !strings.HasPrefix(pkg, ".") {
		return nil, nil, fmt.Errorf("Only relative paths (beginning with a . ) supported for now")
	}
	fset := token.NewFileSet()
	matches, err := filepath.Glob(filepath.Join(pkg, "*.go"))
	if err != nil {
		return nil, nil, err
	}
	files := []*ast.File{}
	for _, match := range matches {
		f, err := parser.ParseFile(fset, match, nil, 0)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, f)
	}
	return fset, files, nil
}

func packageNameToPkg(interfacePackage string) (*types.Package, error) {
	fset, files, err := pkgToFiles(interfacePackage)
	if err != nil {
		return nil, err
	}
	conf := types.Config{Importer: importer.Default()}
	inPkg, err := conf.Check(interfacePackage, fset, files, nil)
	if err != nil {
		return nil, err
	}
	return inPkg, err
}

func parseAndPrintFiles(w io.Writer, interfacePackage, interfaceName, concretePkg, concreteType string) error {
	inPkg, err := packageNameToPkg(interfacePackage)
	if err != nil {
		return err
	}
	if concretePkg == "" {
		concretePkg = interfacePackage
	}
	concPkg := inPkg
	if concretePkg != interfacePackage {
		//only same-package for now, due to import path complexity
		return fmt.Errorf("Only same-package supported")
	}
	err = mix(w, inPkg, concPkg, interfaceName, concreteType)
	return err
}

func parseAndPrint(w io.Writer, input string, interfaceName, concreteType, pkg string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "blah.go", input, 0)
	if err != nil {
		return err
	}
	fs := []*ast.File{f}
	conf := types.Config{Importer: importer.Default()}
	inPkg, err := conf.Check(pkg, fset, fs, nil)
	if err != nil {
		return err
	}
	return mix(w, inPkg, inPkg, interfaceName, concreteType)
}

func mix(w io.Writer, inPkg, concPkg *types.Package, interfaceName string, concreteType string) error {
	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.

	// Print the method sets of Celsius and *Celsius.
	lu := inPkg.Scope().Lookup(interfaceName)
	if lu == nil {
		return fmt.Errorf("Could not find interface '%s'", interfaceName)
	}
	lookupConcType := concPkg.Scope().Lookup(concreteType)
	if lookupConcType != nil {
		return fmt.Errorf("Implementation '%s' already exists", concreteType)
	}

	t := lu.Type()
	//fmt.Fprintf(w,"Method set of %s (type %T):\n", t, t)
	mset := types.NewMethodSet(t)
	for i := 0; i < mset.Len(); i++ {
		//fmt.Fprintf(w,"%#v\n", mset.At(i))

	}
	nu := t.Underlying()

	typeIdentifier := "t"
	//fmt.Fprintf(w,"Underlying type %T):\n", nu)
	k, ok := nu.(*types.Interface)
	if ok {
		//fmt.Println("It's an interface")
		fmt.Fprintf(w, "package %s\n\n", concPkg.Name())
		for _, pkg := range inPkg.Imports() {
			fmt.Fprintf(w, "import \"%s\"\n", pkg.Path())
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "type %s struct {\n\n}\n\n", concreteType)
		for i := 0; i < k.NumMethods(); i++ {
			//	fmt.Fprintf(w,"%#v\n", k.Method(i))
			f := k.Method(i)
			//29.24 vs 34.21 => 20y2m21d down from 26y2m
			//116.23 vs 136.83 => 20y2m21d down from 26y2m
			s, ok := f.Type().(*types.Signature)
			if ok {
				fmt.Fprintf(w, "func (%s *%s) %s(", typeIdentifier, concreteType, f.Name()) //, s.String)
				p := s.Params()
				for i := 0; i < p.Len(); i++ {
					param := p.At(i)
					name := param.Name()
					if name == "" {
						name = "_"
					}
					//paramPkg := param.Pkg().String()
					typedef := param.Type().String()
					fmt.Fprintf(w, "%s %s", name, typedef)
					if i < p.Len()-1 {
						fmt.Fprintf(w, ", ")
					}
				}
				fmt.Fprintf(w, ") (")
				r := s.Results()
				for i := 0; i < r.Len(); i++ {
					restype := r.At(i).Type()
					fmt.Fprintf(w, "%s", restype)
					if i < r.Len()-1 {
						fmt.Fprintf(w, ", ")
					}
				}
				fmt.Fprintf(w, ") {\n")
				if r.Len() == 1 {
					fmt.Fprintf(w, "\treturn nil\n")
				}
			} else {
				fmt.Fprintf(w, "func (%s %s) %s(", typeIdentifier, concreteType, f.Name()) //, s.String)
				fmt.Fprintf(w, "?) {\n")
			}
			//return
			fmt.Fprintf(w, "}\n\n")
		}
	}
	fmt.Println()
	return nil
}
