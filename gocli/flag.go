package gocli

import (
	"flag"
)

type context struct { //Use []string instead of directly touching context
	val *[]string
	raw string
}

func newContext(val *[]string, tags string) *context {
	var err error
	*val, err = splitQuotedFields(tags)
	if err != nil {
		panic(err)
	}
	return &context{val: val, raw: tags}
}

func (c *context) String() string {
	return c.raw
}

func (c *context) IsBoolFlag() bool {
	return false
}

func (c *context) Get() interface{} {
	return *c.val
}

func (c *context) Set(s string) (err error) {
	*c.val, err = splitQuotedFields(s)
	return
}

//TagsFlagSetVar defines a tags flag, matching the tags flag on the go tool.
//
//The tags flag behaves exactly like the -tag flag for the go tool.
//
//The name of the flag is always tags.
//The usage string is set to one similar to the usage of the tags flag
//on the go tool.
func TagsFlagSetVar(f *flag.FlagSet, val *[]string, tags string) {
	fv := newContext(val, tags)
	f.Var(fv, "tags", "a list of build tags. See the documentation of the go/build package for more information about build tags.")
}

//TagsFlagVar defines a tags flag, matching the tags flag on the go tool.
//
//The tags flag behaves exactly like the -tag flag for the go tool.
//
//The name of the flag is always tags.
//The usage string is set to one similar to the usage of the tags flag
//on the go tool.
func TagsFlagVar(val *[]string, tags string) {
	TagsFlagSetVar(flag.CommandLine, val, tags)
}

//TagsFlagSet defines a tags flag, matching the tags flag on the go tool.
//The returned value is the address of a []string variable that stores
//the value of that flag.
//
//The tags flag behaves exactly like the -tag flag for the go tool.
//
//The name of the flag is always tags.
//The usage string is set to one similar to the usage of the tags flag
//on the go tool.
func TagsFlagSet(f *flag.FlagSet, tags string) *[]string {
	var ts []string
	p := &ts
	TagsFlagSetVar(f, p, tags)
	return p
}

//TagsFlag defines a tags flag, matching the tags flag on the go tool.
//The returned value is the address of a []string variable that stores
//the value of that flag.
//
//The tags flag behaves exactly like the -tag flag for the go tool.
//
//The name of the flag is always tags.
//The usage string is set to one similar to the usage of the tags flag
//on the go tool.
func TagsFlag(tags string) *[]string {
	return TagsFlagSet(flag.CommandLine, tags)
}
