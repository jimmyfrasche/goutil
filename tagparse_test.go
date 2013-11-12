package goutil

import (
	"bytes"
	"strings"
	"testing"
)

func bs(s string) []byte {
	return []byte(s)
}

func TestReadUntilPackage(t *testing.T) {
	out := "line one\nline two\n"
	in := strings.NewReader(out + "package")
	lines, err := readUntilPackage(in)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(out, "\n") {
		s := bs(line + "\n")
		if bytes.Equal(lines[i], s) {
			t.Fatal(i, string(lines[i]), "≠", string(s))
		}
	}
}

func mk(s string) (o [][]byte) {
	for _, ln := range strings.Split(s, "\n") {
		o = append(o, bs(ln))
	}
	return
}

var ebt = []struct {
	in  [][]byte
	out []byte
}{
	{
		mk("// +build test"),
		bs("test"),
	},
	{
		mk(`not a comment
// +build test

`),
		bs("test"),
	},
	{
		mk(`// +build test
not a comment`),
		bs("test"),
	},
	{
		mk(`not a comment
// +build test
also not a comment`),
		bs("test"),
	},
	{
		mk("// +build a\n// +build b"),
		bs("a b"),
	},
}

func TestExtractBuildTags(t *testing.T) {
	for i, tc := range ebt {
		bt := extractBuildTags(tc.in)
		if !bytes.Equal(bt, tc.out) {
			t.Error(i, string(bt), "≠", string(tc.out))
		}
	}
}

//for the sake of simplicity we unjustly assume that that order is important.
func tagcmp(a, b tag) bool {
	if a == nil {
		return b == nil
	}

	switch ap := a.(type) {
	case atag:
		if bp, ok := b.(atag); ok {
			return string(ap) == string(bp)
		}
	case negtag:
		if bp, ok := b.(negtag); ok {
			return string(ap) == string(bp)
		}
	case andtag:
		if bp, ok := b.(andtag); ok {
			if len(ap) != len(bp) {
				return false
			}
			for i := range ap {
				if ap[i] == nil || bp[i] == nil {
					return false
				}
				if !tagcmp(ap[i], bp[i]) {
					return false
				}
			}
			return true
		}
	case ortag:
		if bp, ok := b.(ortag); ok {
			if len(ap) != len(bp) {
				return false
			}
			for i := range ap {
				if ap[i] == nil || bp[i] == nil {
					return false
				}
				if !tagcmp(ap[i], bp[i]) {
					return false
				}
			}
			return true
		}
	}

	return false
}

func tagfmt(t tag) string {
	if t == nil {
		return ""
	}

	var acc []string
	push := func(t tag) {
		acc = append(acc, tagfmt(t))
	}
	joiner := " "
	switch t := t.(type) {
	case atag:
		return string(t)
	case negtag:
		return "!" + string(t)
	case ortag:
		for _, o := range t {
			push(o)
		}
	case andtag:
		joiner = ","
		for _, a := range t {
			push(a)
		}
	}
	return strings.Join(acc, joiner)
}

var te = []struct {
	in  string
	out tag
}{
	{"", nil},
	{"tag", atag("tag")},
	{"!tag", negtag("tag")},
	{"a b", ortag{atag("a"), atag("b")}},
	{"a b c", ortag{atag("a"), atag("b"), atag("c")}},
	{"a,b", andtag{atag("a"), atag("b")}},
	{"a,b,c", andtag{atag("a"), atag("b"), atag("c")}},
	{"a b,c", ortag{atag("a"), andtag{atag("b"), atag("c")}}},
	{"a,b c", ortag{andtag{atag("a"), atag("b")}, atag("c")}},
	{
		"a,!b !c,d",
		ortag{
			andtag{atag("a"), negtag("b")},
			andtag{negtag("c"), atag("d")},
		},
	},
}

func TestParse(t *testing.T) {
	for i, io := range te {
		tag := parseOr(bs(io.in))
		if !tagcmp(tag, io.out) {
			t.Log(i, io.in, tagfmt(io.out))
			t.Logf("%#v\n", tag)
			t.Errorf("%#v\n", io.out)
		}
	}
}

var tm = []struct {
	matches bool
	what    []string
	with    tag
}{}

func TestMatch(t *testing.T) {
	for i, m := range tm {
		if m.matches != m.with.match(m.what) {
			t.Errorf("%d fails: %#v\n", i, m)
		}
	}
}
