package tree

import (
	"fmt"
	"go/build"
	"io"
	"strings"

	"go.mrchanchal.com/treewriter"
)

var _ fmt.Stringer = (*Package)(nil)

type Package struct {
	Name string
	Doc  string
}

func (p *Package) String() string {
	if p.Doc == "" {
		return p.Name
	}

	nameS := strings.Split(p.Name, "/")
	doc := strings.TrimPrefix(p.Doc, "Package ")
	doc = strings.TrimPrefix(doc, nameS[len(nameS)-1]+" ")
	doc = strings.TrimSuffix(doc, ".")

	return fmt.Sprintf("%s [%s]", p.Name, doc)
}

var (
	_ treewriter.Tree = (*Tree)(nil)
	_ fmt.Stringer    = (*Tree)(nil)
	_ io.WriterTo     = (*Tree)(nil)
)

type Tree struct {
	Name        string
	Doc         string
	Childs      []*Tree
	PackageList []*Package
}

func (t *Tree) GetPackage() *Package {
	return &Package{
		Name: t.Name,
		Doc:  t.Doc,
	}
}

func (t *Tree) WriteTo(writer io.Writer) (int64, error) {
	return treewriter.WriteTo(writer, t, nil)
}

func (t *Tree) Children() []treewriter.Tree {
	childs := make([]treewriter.Tree, len(t.Childs))

	for i, child := range t.Childs {
		childs[i] = child
	}

	return childs
}

func (t *Tree) String() string {
	return t.Name
}

func ImportTree(path, srcDir string) (*Tree, error) {
	tree, packageMap, err := importTree(path, srcDir, make(map[string]*Tree))
	if err != nil {
		return nil, err
	}

	tree.PackageList = make([]*Package, 0, len(packageMap))
	for _, pkg := range packageMap {
		tree.PackageList = append(tree.PackageList, pkg.GetPackage())
	}

	return tree, nil
}

func importTree(path, srcDir string, cache map[string]*Tree) (*Tree, map[string]*Tree, error) {
	if path == "C" {
		return nil, cache, nil
	}

	if cached, ok := cache[path]; ok {
		return cached, cache, nil
	}

	pkg, err := build.Import(path, srcDir, 0)
	if err != nil {
		return nil, cache, err
	}

	if pkg.Goroot {
		return nil, cache, nil
	}

	imports := make([]*Tree, 0)

	for _, imp := range pkg.Imports {
		pkg, tempCache, err := importTree(imp, srcDir, cache)
		if err != nil {
			return nil, tempCache, err
		}

		if pkg != nil {
			imports = append(imports, pkg)
			cache = tempCache
		}
	}

	tree := Tree{
		Name:        pkg.ImportPath,
		Doc:         pkg.Doc,
		Childs:      imports,
		PackageList: nil,
	}

	cache[tree.Name] = &tree

	return &tree, cache, nil
}
