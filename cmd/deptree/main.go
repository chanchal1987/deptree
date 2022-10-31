package main

import (
	"flag"
	"fmt"
	"os"

	deptree "go.mrchanchal.com/deptree"
	sort "go.mrchanchal.com/slicesort"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func must[T any](val T, err error) T {
	panicOnErr(err)

	return val
}

func main() {
	writeTo := os.Stdout

	flag.Func("f", "write to a file instead of STDOUT", func(name string) error {
		f, err := os.Create(name)
		if err != nil {
			return err
		}

		writeTo = f

		return nil
	})

	defer func() { panicOnErr(writeTo.Close()) }()

	flag.Parse()

	var path, srcDir string

	switch flag.NArg() {
	case 0:
		path = "."
		srcDir = must(os.Getwd())
	case 1:
		path = flag.Arg(0)
		srcDir = must(os.Getwd())
	default:
		path = flag.Arg(0)
		srcDir = flag.Arg(1)
	}

	tree := must(deptree.ImportTree(path, srcDir))
	must(fmt.Fprintln(writeTo, "Packages Found:"))

	m := tree.PackageList
	sort.Sort(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})

	for _, pkg := range m {
		must(fmt.Fprintf(writeTo, "%v\n", pkg))
	}

	must(fmt.Fprintln(writeTo, "\nPackage Tree:"))
	must(tree.WriteTo(writeTo))
}
