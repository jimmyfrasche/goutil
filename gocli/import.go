package gocli

import (
	"errors"
	"go/build"
	"path/filepath"

	"github.com/jimmyfrasche/goutil"
)

//just a basic wrapper to give goutil.Import the same signature as ImportTree
func imports(ctx *build.Context, ip string) (pkgs goutil.Packages, err error) {
	pkg, err := goutil.Import(ctx, ip)
	if err != nil {
		return nil, err
	}
	return append(pkgs, pkg), nil
}

//Import is for importing command line arguments.
//It uses the following rules:
//	If args is len 0, try to import the current directory.
//	Otherwise, for each argument:
//		If it ends with ..., use goutil.ImportTree (unless notree is true)
//		Otherwise, use goutil.Import
//Regardless, the ctx is passed as is to the various importers.
//
//Import returns a list of any errors that resulted from attmepting to import.
//If you only care about the first error, wrap the call in FirstError.
func Import(notree bool, ctx *build.Context, args []string) (pkgs []goutil.Packages, errs []error) {
	push := func(ps goutil.Packages, err error) {
		pkgs = append(pkgs, ps)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(args) == 0 {
		push(imports(ctx, ""))
		return
	}

	for _, arg := range args {
		if filepath.Base(arg) == "..." {
			if notree {
				push(nil, errors.New("cannot use ... imports"))
			} else {
				push(goutil.ImportTree(ctx, filepath.Dir(arg)))
			}
		} else {
			push(imports(ctx, arg))
		}
	}
	return
}

//FirstError is meant to wrap Import, it returns pkgs unchanged.
//If there were errors, only the first is returned.
func FirstError(pkgs []goutil.Packages, errs []error) ([]goutil.Packages, error) {
	if len(errs) == 0 {
		return pkgs, nil
	}
	return pkgs, errs[0]
}

//Flatten takes a slice of goutil.Packages and returns a single goutil.Packages
//containing only the unique packages from the slice.
func Flatten(pss []goutil.Packages) (out goutil.Packages) {
	for _, ps := range pss {
		out = append(out, ps...)
	}
	return out.Uniq()
}

//ImportOne only allows one argument to be specified.
func ImportOne(notree bool, ctx *build.Context, args []string) (goutil.Packages, error) {
	if len(args) > 1 {
		return nil, errors.New("only one package may be specified")
	}
	ps, err := FirstError(Import(notree, ctx, args))
	if err != nil {
		return nil, err
	}
	return ps[0], nil
}
