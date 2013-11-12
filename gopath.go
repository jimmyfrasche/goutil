package goutil

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

var gopaths = build.Default.SrcDirs()

func init() {
	fs := string(filepath.Separator)
	//so we can use strings.HasPrefix cleanly
	for i, p := range gopaths {
		gopaths[i] = filepath.Clean(p) + fs
	}
}

func isdir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

//ToImport takes an arbitrary path and returns a valid import
//path using $GOPATH. Returns which $GOPATH and the import path.
//
//ToImport iterates $GOPATH in the order it's defined and always returns
//the first match.
//
//Works correctly with imports in the Go root.
//
//Does not handle the special glob ... (see ImportTree).
//
//ToImport handles relative paths and "." The empty string is treated
//the same as "."
func ToImport(path string) (root, imp string, err error) {
	op := path //save original for error reporting
	if path == "" || path == "." {
		op = "."
		path, err = os.Getwd()
		if err != nil {
			return
		}
	}
	path = filepath.Clean(path)

	//see if path is absolute
	for _, p := range gopaths {
		if strings.HasPrefix(path, p) {
			if isdir(p) {
				return p, path[len(p):], nil
			}
		}
	}

	//just given an import path
	for _, p := range gopaths {
		if isdir(filepath.Join(p, path)) {
			return p, path, nil
		}
	}

	return "", "", fmt.Errorf("Directory %s not in $GOPATH", op)
}
