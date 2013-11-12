package goutil

import (
	"go/ast"
	"go/token"
	"strings"
)

//Decls is a list ast.Decls.
type Decls []ast.Decl

//Decls returns every ast.Decl in a package that is not a BadDecl or an IMPORT.
//
//It is up to the caller to call Parse before invoking this method.
func (p *Package) Decls() (ds Decls) {
	for _, f := range p.AST.Files {
		for _, d := range f.Decls {
			switch dt := d.(type) {
			case *ast.GenDecl:
				if dt.Tok != token.IMPORT {
					ds = append(ds, d)
				}
			case *ast.FuncDecl:
				ds = append(ds, d)
			}
		}
	}
	return
}

//Funcs returns a Decls filtered to just *ast.FuncDecl
func (ds Decls) Funcs() (out Decls) {
	for _, d := range ds {
		if _, ok := d.(*ast.FuncDecl); ok {
			out = append(out, d)
		}
	}
	return
}

func (ds Decls) gendecl(t token.Token) (out Decls) {
	for _, d := range ds {
		if dt, ok := d.(*ast.GenDecl); ok {
			if dt.Tok == t {
				out = append(out, d)
			}
		}
	}
	return
}

//Types returns a Decls filtered to just *ast.GenDecl containing types.
func (ds Decls) Types() Decls {
	return ds.gendecl(token.TYPE)
}

//Consts returns a Decls filtered to just *ast.GenDecl containing consts.
func (ds Decls) Consts() Decls {
	return ds.gendecl(token.CONST)
}

//Vars returns a Decls filtered to just *ast.GenDecl containing vars.
func (ds Decls) Vars() Decls {
	return ds.gendecl(token.VAR)
}

func dupdecl(g *ast.GenDecl, s ast.Spec) *ast.GenDecl {
	return &ast.GenDecl{
		Doc:    g.Doc,
		TokPos: g.TokPos,
		Tok:    g.Tok,
		Lparen: g.Lparen,
		Specs:  []ast.Spec{s},
		Rparen: g.Rparen,
	}
}

//SplitSpecs goes throught each *ast.GenDecl and duplicates the outer GenDecl
//for each item in the spec. This destroys information but does make it easier
//to search, and is required for invoking Named on GenDecls.
//
//Note that SplitSpecs does not, and in general cannot sensibly, split
//	var a, b, c = f()
//though it does split
//	var a, b, c = 1, 2, 3
//as one would hope.
func (ds Decls) SplitSpecs() (out Decls) {
	for _, d := range ds {
		if g, ok := d.(*ast.GenDecl); ok {
			for _, s := range g.Specs {
				switch st := s.(type) {
				case *ast.ValueSpec:
					if len(st.Names) > 1 && len(st.Names) == len(st.Values) {
						for i := range st.Names {
							out = append(out, dupdecl(g, &ast.ValueSpec{
								Doc:     st.Doc,
								Names:   []*ast.Ident{st.Names[i]},
								Type:    st.Type,
								Values:  []ast.Expr{st.Values[i]},
								Comment: st.Comment,
							}))
						}
					} else {
						out = append(out, dupdecl(g, s))
					}
				case *ast.TypeSpec:
					out = append(out, dupdecl(g, s))
				}
			}
		} else {
			out = append(out, d)
		}
	}
	return
}

//StringMatcher matches strings.
//
//The interface is extracted from regexp.Regexp.
type StringMatcher interface {
	MatchString(string) bool
}

//PrefixMatcher matches all strings with specified prefix.
type PrefixMatcher string

//MatchString matches strings with the prefix set
func (p PrefixMatcher) MatchString(s string) bool {
	return strings.HasPrefix(s, string(p))
}

//Named returns all Decls whose name matches r.
//
//If you haven't called SplitSpecs, a GenDecl will be returned
//if any of its Spec's match.
func (ds Decls) Named(m StringMatcher) (out Decls) {
	for _, d := range ds {
		switch dt := d.(type) {
		case *ast.FuncDecl:
			if m.MatchString(dt.Name.Name) {
				out = append(out, d)
			}
		case *ast.GenDecl:
			matched := false
			for _, s := range dt.Specs {
				switch st := s.(type) {
				case *ast.TypeSpec:
					if m.MatchString(st.Name.Name) {
						matched = true
					}
				case *ast.ValueSpec:
					for _, nm := range st.Names {
						if m.MatchString(nm.Name) {
							matched = true
						}
					}
				}
			}
			if matched {
				out = append(out, d)
			}
		}
	}
	return
}
