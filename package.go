package goutil

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

//Package wraps all the common go/* Package types and provides some
//additional functionality.
//
//All methods of package assume that the Package was created
//by either Import or ImportAll and hence that Build is nonnil.
type Package struct {
	//The context this Package was imported with.
	Context *build.Context
	Build   *build.Package
	AST     *ast.Package
	Doc     *doc.Package
	//filename → tag
	tags map[string]tag
}

//ParseTags parses the build tags for each file in Build.GoFiles.
//
//ParseTags uses its own parser so it is not necessary to call Parse
//before calling ParseTags.
func (p *Package) ParseTags() error {
	tags := map[string]tag{}
	for _, f := range p.Build.GoFiles {
		file, err := os.Open(filepath.Join(p.Build.Dir, f))
		if err != nil {
			return err
		}
		bt, err := parseTags(file)
		if err != nil {
			return err
		}
		if bt != nil {
			tags[f] = bt
		}
		file.Close()
	}
	p.tags = tags
	return nil
}

//FilesMatching returns the sublist of Build.GoFiles matching the specified
//build tags.
//
//It is the users responsibility to call ParseTags before invoking
//this method.
func (p *Package) FilesMatching(tags ...string) (files []string) {
	if len(tags) == 0 {
		return nil
	}
	for file, tag := range p.tags {
		if tag.match(tags) {
			files = append(files, file)
		}
	}
	return
}

//ASTFilesMatching calls FilesMatching and looks up the resulting files
//in p.AST.Files.
//
//It is the callers responsibility to call Parse and ParseTags before
//invoking this method.
func (p *Package) ASTFilesMatching(tags ...string) (files []*ast.File) {
	for _, file := range p.FilesMatching(tags...) {
		files = append(files, p.AST.Files[file])
	}
	return
}

func (p *Package) parse(pc bool) (*ast.Package, error) {
	f := func(fi os.FileInfo) bool {
		if fi.IsDir() {
			return false
		}
		nm := fi.Name()
		for _, f := range p.Build.GoFiles {
			if nm == f {
				return true
			}
		}
		return false
	}

	var m parser.Mode
	if pc {
		m = parser.ParseComments
	}

	pkgs, err := parser.ParseDir(token.NewFileSet(), p.Build.Dir, f, m)
	if err != nil {
		return nil, err
	}

	pkg, ok := pkgs[p.Build.Name]
	if !ok {
		//I do not even know how this could happen but may as well handle it
		//in case any assumptions shift from under our feet.
		return nil, fmt.Errorf("No package named %s", p.Build.Name)
	}
	return pkg, nil
}

//Parse the package and set p.AST.
//
//It is not necessary to call with parseComments if you intend to call
//ParseDocs, as ParseDocs creates its own parse.
func (p *Package) Parse(parseComments bool) error {
	if p.AST != nil {
		return nil
	}

	pkg, err := p.parse(parseComments)
	if err != nil {
		return err
	}

	p.AST = pkg
	return nil
}

//ParseDocs parses the Package's documentation with go/doc.
//
//If you do not need a particular doc.Mode call this with 0.
//
//If the package directory contains a file of package documentation
//(and the package is not itself named documentation), it is parsed
//and its doc.Package.Doc string replaces the string generated
//by the package itself.
//
//Note that the go/doc package munges the AST so this method parses the AST
//again, regardless of the value in p.AST. As a consequence, it is valid
//to call this even if you have not called the Parse method or if you have
//called the Parse method and told it not to parse comments.
func (p *Package) ParseDocs(mode doc.Mode) error {
	if p.Doc != nil {
		return nil
	}
	pkg, err := p.parse(true)
	if err != nil {
		return err
	}

	p.Doc = doc.New(pkg, p.Build.ImportPath, mode)

	//we don't want the below running if we happen to be importing a package
	//whose name happens to be documentation.
	if p.Build.Name == "documentation" {
		return nil
	}

	//check ignored files for any package named documentation.
	//assume there is only one such file.
	//We ignore errors here as the ignored files may not be meant to parse.
	var docfile string
	for _, u := range p.Build.IgnoredGoFiles {
		path := filepath.Join(p.Build.Dir)
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, path, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		if f.Name.Name == "documentation" {
			docfile = u
			break
		}
	}

	//there's an ignored file of package documentation,
	//parse it and replace the package doc string with this doc string.
	if docfile != "" {
		fs := token.NewFileSet()
		f := func(fi os.FileInfo) bool {
			return !fi.IsDir() && fi.Name() == docfile
		}
		pkgs, err := parser.ParseDir(fs, p.Build.Dir, f, parser.ParseComments)
		if err != nil {
			return err
		}
		d := doc.New(pkgs["documentation"], p.Build.ImportPath, 0)
		p.Doc.Doc = d.Doc
	}

	return nil
}
