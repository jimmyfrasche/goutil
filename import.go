package goutil

import (
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

type ident struct {
	ctx *build.Context
	imp string
}

var (
	pkgcache   = map[ident]*Package{}
	cmux       = new(sync.Mutex)
	defaultctx = &build.Default
)

//Context returns a *build.Context with the appropriate tags.
//
//This does not change GOARCH or GOOS, it only sets additional tags.
//
//If no tags are specified, the default context is returned.
func Context(tags ...string) *build.Context {
	if len(tags) == 0 {
		return defaultctx
	}
	c := build.Default
	//BuildTags currently always nil but nothing says that can't change.
	c.BuildTags = append(c.BuildTags, tags...)
	return &c
}

func pkgget(ident ident) *Package {
	cmux.Lock()
	defer cmux.Unlock()
	return pkgcache[ident]
}

func pkgset(ident ident, pkg *Package) {
	cmux.Lock()
	defer cmux.Unlock()
	pkgcache[ident] = pkg
}

//Import imports a package.
//
//path is run through ToImport.
//
//If ctx is nil, the default context is used.
//
//N.B. we require a pointer to a build.Context for caching.
//Two build contexts with identical values that are not represented
//by the same pointer will have all packages imported by them
//cached separately.
func Import(ctx *build.Context, path string) (*Package, error) {
	if ctx == nil {
		ctx = defaultctx
	}
	root, path, err := ToImport(path)
	if err != nil {
		return nil, err
	}

	ident := ident{ctx, path}
	if pkg := pkgget(ident); pkg != nil {
		return pkg, nil
	}

	p, err := ctx.Import(path, root, 0)
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		Context: ctx,
		Build:   p,
	}
	pkgset(ident, pkg)

	return pkg, nil
}

//ImportTree imports every package in the directory tree rooted at root.
//Root need not have a valid package.
//
//If there are any errors, the first is returned and as many packages as can be
//imported are returned.
func ImportTree(ctx *build.Context, root string) (pkgs Packages, first error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	importdir(ctx, root, &pkgs, &first)

	return
}

//ImportAll imports every package from the standard library
//and every $GOPATH.
//
//If there are any errors, the first is returned and as many packages as can be
//imported are returned.
func ImportAll(ctx *build.Context) (pkgs Packages, first error) {
	for _, root := range gopaths {
		p, err := ImportTree(ctx, root)
		if first != nil {
			first = err
		}
		pkgs = append(pkgs, p...)
	}
	return
}

func importdir(ctx *build.Context, root string, acc *Packages, first *error) {
	hasGoFiles := false
	var subdirs []string

	fis, err := ioutil.ReadDir(root)
	if err != nil && *first == nil {
		*first = err
	}

	for _, fi := range fis {
		nm := fi.Name()
		if fi.IsDir() {
			subdirs = append(subdirs, nm)
		} else if strings.HasSuffix(nm, ".go") {
			hasGoFiles = true
		}
	}

	if hasGoFiles {
		pkg, err := Import(ctx, root)
		if err != nil && *first == nil {
			*first = err
		} else {
			*acc = append(*acc, pkg)
		}
	}

	for _, dir := range subdirs {
		importdir(ctx, filepath.Join(root, dir), acc, first)
	}
}

//ImportDeps imports all dependencies of this package recursively.
//
//It uses the same build.Context this Package was built with.
func (p *Package) ImportDeps() (pkgs Packages, err error) {
	seen := map[string]bool{p.Build.ImportPath: true, "C": true}
	acc := Packages{p}
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				pkgs, err = nil, e
			} else {
				panic(x)
			}
		}
	}()
	recimp(p.Context, p, &acc, seen)
	return acc, nil
}

func recimp(ctx *build.Context, root *Package, acc *Packages, seen map[string]bool) {
	for _, ch := range root.Build.Imports {
		if !seen[ch] {
			seen[ch] = true
			p, err := Import(ctx, ch)
			if err != nil {
				panic(err)
			}
			*acc = append(*acc, p)
			recimp(ctx, p, acc, seen)
		}
	}
}

//ImportRec calls Import on path and then ImportDeps and returns all packages
//with the package described by path as the first element.
func ImportRec(ctx *build.Context, path string) (Packages, error) {
	pkg, err := Import(ctx, path)
	if err != nil {
		return nil, err
	}

	pkgs, err := pkg.ImportDeps()
	if err != nil {
		return nil, err
	}

	return append(Packages{pkg}, pkgs...), nil
}
