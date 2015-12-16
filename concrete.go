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
	"log"
	"os"
	"text/template"
)

var (
	interfaceName   = flag.String("interface", "", "The name of the interface to be implemented")
	filename        = flag.String("in-file", "", "file where interface is defined")
	implPackage     = flag.String("impl-package", "", "package")
	concreteTypeTpl = flag.String("implementation", "{{.Interface}}Impl", "The name of the concrete implementation")
	helpFlag        = flag.Bool("help", false, "show detailed help message")
	writeFlag       = flag.Bool("w", false, "rewrite input files in place (by default, the results are printed to standard output)")
	verboseFlag     = flag.Bool("v", false, "show verbose matcher diagnostics")
)

const usage = `concrete: a tool to implement interfaces

Usage: concrete -interface <InterfaceName> -in-file <existing-file.go> -impl-package <package> [options]

-in-file         existing file which contains the interface
-interface       name of interface
-impl-package    package name of existing interface
-concrete        The name of the implementation. Uses templates. default '{{.Interface}}Impl'
-help            show detailed help message
`

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "eg: %s\n", err)
		os.Exit(1)
	}
}

func doMain() error {
	flag.Parse()

	if *helpFlag || *filename == "" {
		fmt.Fprint(os.Stderr, usage)
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
	parseAndPrintFile(*filename, *interfaceName, out.String(), *implPackage)
	return nil
}

func parseAndPrintFile(filename, interfaceName, concreteType, pkg string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	mix(fset, f, interfaceName, concreteType, pkg)
}

func parseAndPrint(input string, interfaceName, concreteType, pkg string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "blah.go", input, 0)
	if err != nil {
		log.Fatal(err)
	}
	mix(fset, f, interfaceName, concreteType, pkg)
}

func mix(fset *token.FileSet, f *ast.File, interfaceName string, concreteType string, pkgName string) {
	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(pkgName, fset, []*ast.File{f}, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Print the method sets of Celsius and *Celsius.
	t := pkg.Scope().Lookup(interfaceName).Type()
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
		fmt.Printf("package %s\n\n", pkgName)
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
			} else {
				fmt.Printf("func (%s %s) %s(", typeIdentifier, concreteType, f.Name()) //, s.String)
				fmt.Printf("?) {\n")
			}
			//return
			fmt.Printf("}\n\n")
		}
	}
	fmt.Println()

}
