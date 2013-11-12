package goutil

import (
	"go/doc"
)

//Packages is a list of *Package, with methods for filtering and manipulating
//en masse.
type Packages []*Package

//Parse parses each package in turn.
//
//Parse stops and returns the first error it encounters.
func (ps Packages) Parse(parseComments bool) (err error) {
	for _, p := range ps {
		if err = p.Parse(parseComments); err != nil {
			break
		}
	}
	return
}

//ParseDocs parses each package's documentation in turn.
//
//ParseDocs stops and returns the first error it encounters.
func (ps Packages) ParseDocs(mode doc.Mode) (err error) {
	for _, p := range ps {
		if err = p.ParseDocs(mode); err != nil {
			break
		}
	}
	return
}

//ParseTags parses each package's build tags in turn.
//
//ParseTags stops and returns the first error it encounters.
func (ps Packages) ParseTags() (err error) {
	for _, p := range ps {
		if err = p.ParseTags(); err != nil {
			break
		}
	}
	return
}

//Filter returns a sublist of packages that match the predicate f.
//
//If the predicate requires the Packages to be parsed or have their docs
//parsed, it is up to the user to ensure that these steps have been done first.
func (ps Packages) Filter(f func(*Package) bool) (out Packages) {
	for _, p := range ps {
		if f(p) {
			out = append(out, p)
		}
	}
	return
}

//NoStdlib filters out packages from GOROOT.
//
//This only relies on information in Build, so it is safe to call
//if the packages have not been parsed.
func (ps Packages) NoStdlib() Packages {
	return ps.Filter(func(p *Package) bool {
		return !p.Build.Goroot
	})
}

//HasFilesMatching returns all packages that have at least one file matching
//the specified build tags.
//
//HasFilesMatching requires that ParseTags has been invoked.
func (ps Packages) HasFilesMatching(tags ...string) Packages {
	return ps.Filter(func(p *Package) bool {
		return len(p.FilesMatching(tags...)) > 0
	})
}
