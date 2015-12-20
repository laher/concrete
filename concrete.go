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
				fmt.Printf(" %s\n", lu.Name())
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
		fmt.Printf("Error parsing template")
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
		fmt.Printf("Error processing template")
		return err
	}
	return parseAndPrintFiles(*inPackage, *interfaceName, *implPackage, out.String())
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

func parseAndPrintFiles(interfacePackage, interfaceName, concretePkg, concreteType string) error {
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
	err = mix(inPkg, concPkg, interfaceName, concreteType)
	return err
}

func parseAndPrint(input string, interfaceName, concreteType, pkg string) error {
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
	return mix(inPkg, inPkg, interfaceName, concreteType)
}

func mix(inPkg, concPkg *types.Package, interfaceName string, concreteType string) error {
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
	//fmt.Printf("Method set of %s (type %T):\n", t, t)
	mset := types.NewMethodSet(t)
	for i := 0; i < mset.Len(); i++ {
		//fmt.Printf("%#v\n", mset.At(i))

	}
	nu := t.Underlying()

	typeIdentifier := "t"
	//fmt.Printf("Underlying type %T):\n", nu)
	k, ok := nu.(*types.Interface)
	if ok {
		//fmt.Println("It's an interface")
		fmt.Printf("package %s\n\n", concPkg.Name())
		for _, pkg := range inPkg.Imports() {
			fmt.Printf("import \"%s\"\n", pkg.Path())
		}
		fmt.Printf("\n")
		fmt.Printf("type %s struct {\n\n}\n\n", concreteType)
		for i := 0; i < k.NumMethods(); i++ {
			//	fmt.Printf("%#v\n", k.Method(i))
			f := k.Method(i)
			//29.24 vs 34.21 => 20y2m21d down from 26y2m
			//116.23 vs 136.83 => 20y2m21d down from 26y2m
			s, ok := f.Type().(*types.Signature)
			if ok {
				fmt.Printf("func (%s *%s) %s(", typeIdentifier, concreteType, f.Name()) //, s.String)
				p := s.Params()
				for i := 0; i < p.Len(); i++ {
					param := p.At(i)
					name := param.Name()
					if name == "" {
						name = "_"
					}
					//paramPkg := param.Pkg().String()
					typedef := param.Type().String()
					fmt.Printf("%s %s", name, typedef)
					if i < p.Len()-1 {
						fmt.Printf(", ")
					}
				}
				fmt.Printf(") (")
				r := s.Results()
				for i := 0; i < r.Len(); i++ {
					restype := r.At(i).Type()
					fmt.Printf("%s", restype)
					if i < r.Len()-1 {
						fmt.Printf(", ")
					}
				}
				fmt.Printf(") {\n")
				if r.Len() == 1 {
					fmt.Printf("\treturn nil\n")
				}
			} else {
				fmt.Printf("func (%s %s) %s(", typeIdentifier, concreteType, f.Name()) //, s.String)
				fmt.Printf("?) {\n")
			}
			//return
			fmt.Printf("}\n\n")
		}
	}
	fmt.Println()
	return nil
}
