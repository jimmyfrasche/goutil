package goutil

import (
	"go/build"
)

//TagsOf returns the complete build tag specification
//of a build.Context.
func TagsOf(c *build.Context) []string {
	return append(c.BuildTags, c.GOOS, c.GOARCH)
}
