//Declgrep greps just the declarations, exported and unexported,
//excluding import declarations, of a given Go package,
//or set of Go packages, with the standard build tags.
//
//Declgrep runs an RE2-style regular expression against the names
//of the declarations in one or more packages. The packages may be
//specified as with the go(1) tool, including the special ...
//operator. The -r flag searches both the specified package and
//its dependencies, even when invoked with the ... operator.
//
//The line number in the output is not guaranteed to be exact,
//except in the case of functions. Exported and unexported
//declarations are searched.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go/ast"
	"go/printer"
	"go/token"

	"github.com/jimmyfrasche/goutil"
)

var (
	r        = flag.Bool("r", false, "recursively search dependencies")
	v        = flag.Bool("v", false, "select non-matching declarations")
	l        = flag.Bool("l", false, "prefer leftmost-longest matches")
	nostdlib = flag.Bool("nostdlib", false, "do not match against standard library")
)

//invert regex matches for -v
type negmatcher struct {
	re *regexp.Regexp
}

func (n negmatcher) MatchString(s string) bool {
	return !n.re.MatchString(s)
}

func Usage() {
	_, nm := filepath.Split(os.Args[0])
	log.Printf("Usage: %s [flags] regexp [package|directory]\n", nm)
	flag.PrintDefaults()
}

//use go/printer to print tiny expressions. If more than one line, fix.
func fmtast(fs *token.FileSet, v interface{}) string {
	var b bytes.Buffer
	printer.Fprint(&b, fs, v)
	out := strings.SplitN(b.String(), "\n", 2)
	s := out[0]
	if len(out) > 1 {
		s += " ..."
	}
	return s
}

func fmtpos(pkg *goutil.Package, pos token.Pos) string {
	s := ""
	p := pkg.FileSet.Position(pos)
	_, f := filepath.Split(p.Filename)
	return s + fmt.Sprintf("%s:%d:", f, p.Line)
}

//print just types, caller handles ()
func fmtlist(fs *token.FileSet, fields []*ast.Field) string {
	var acc []string
	for _, f := range fields {
		t := fmtast(fs, f.Type)
		for _ = range f.Names {
			acc = append(acc, t)
		}
	}
	return strings.Join(acc, ", ")
}

func fmtfunc(fs *token.FileSet, f *ast.FuncDecl) string {
	name := f.Name.Name

	method := ""
	if f.Recv != nil && len(f.Recv.List) > 0 {
		method = " (" + fmtlist(fs, f.Recv.List) + ")"
	}

	t := f.Type
	params := "(" + fmtlist(fs, t.Params.List) + ")"

	ret := ""
	if t.Results != nil {
		if nr := len(t.Results.List); nr > 0 {
			ret = fmtlist(fs, t.Results.List)
			if nr > 1 || len(t.Results.List[0].Names) > 1 {
				ret = "(" + ret + ")"
			}
		}
		ret = " " + ret
	}

	return fmt.Sprintf("func%s %s%s%s", method, name, params, ret)
}

func print(showimp bool, p *goutil.Package, d ast.Decl) {
	where, what := "", ""
	if showimp {
		where = p.Build.ImportPath + ":"
	}

	switch dt := d.(type) {
	case *ast.FuncDecl:
		where += fmtpos(p, dt.Type.Func)
		what = fmtfunc(p.FileSet, dt)

	case *ast.GenDecl:
		where += fmtpos(p, dt.TokPos)
		what = dt.Tok.String() + " " + fmtast(p.FileSet, dt.Specs[0])
	}
	fmt.Println(where, what)
}

//Usage: %name %flags regexp [package|directory]
func main() {
	log.SetFlags(0)
	fatal := log.Fatalln

	flag.Usage = Usage
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || len(args) > 2 {
		Usage()
		os.Exit(2)
	}

	re, err := regexp.Compile(args[0])
	if err != nil {
		fatal(err)
	}
	if *l {
		re.Longest()
	}

	var m goutil.StringMatcher = re
	if *v {
		m = negmatcher{re}
	}

	tree := false
	imp := "."
	if len(args) > 1 {
		imp = args[1]
		if d, f := filepath.Split(imp); f == "..." {
			tree = true
			imp = d
		}
	}

	var pkgs goutil.Packages
	switch {
	case tree && *r:
		pkgs, err := goutil.ImportTree(nil, imp)
		if err != nil {
			log.Println(err)
		}
		var ps goutil.Packages
		for _, p := range pkgs {
			t, err := p.ImportDeps()
			if err != nil {
				fatal(err)
			}
			ps = append(ps, t...)
		}
		pkgs = append(pkgs, ps...).Uniq()
	case tree:
		pkgs, err = goutil.ImportTree(nil, imp)
		if err != nil {
			log.Println(err)
		}
	case *r:
		pkgs, err = goutil.ImportRec(nil, imp)
		if err != nil {
			fatal(err)
		}
	default:
		p, err := goutil.Import(nil, imp)
		if err != nil {
			fatal(err)
		}
		pkgs = goutil.Packages{p}
	}

	if *nostdlib {
		pkgs = pkgs.NoStdlib()
	}

	err = pkgs.Parse(false)
	if err != nil {
		fatal(err)
	}

	multiples := len(pkgs) > 1
	for _, pkg := range pkgs {
		for _, d := range pkg.Decls().SplitSpecs().Named(m) {
			print(multiples, pkg, d)
		}
	}
}
