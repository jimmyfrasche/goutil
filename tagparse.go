package goutil

import (
	"bufio"
	"bytes"
	"io"
)

func readUntilPackage(f io.Reader) (lines [][]byte, err error) {
	scanner := bufio.NewScanner(f)
	pkg := []byte("package ")
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if bytes.HasPrefix(line, pkg) {
			break
		}
		lines = append(lines, line)
	}
	if err = scanner.Err(); err != nil {
		lines = nil
	}
	return
}

//returns all lines that are build tags minus "// +build"
func extractBuildTags(lines [][]byte) []byte {
	buildline := []byte("+build ")
	inbuild := false
	var out [][]byte
	for _, line := range lines {
		p := bytes.Index(line, buildline)
		if p < 0 {
			if inbuild {
				break
			}
			continue
		} else {
			inbuild = true
		}
		out = append(out, bytes.TrimSpace(line[p+len(buildline):]))
	}
	return bytes.Join(out, []byte{' '})
}

type tag interface {
	match(tags []string) bool
}

type atag string

func (t atag) match(tags []string) bool {
	for _, tag := range tags {
		if string(t) == tag {
			return true
		}
	}
	return false
}

type negtag string

func (n negtag) match(tags []string) bool {
	return !atag(string(n)).match(tags)
}

type andtag []tag

func (a andtag) match(tags []string) bool {
	for _, t := range a {
		if !t.match(tags) {
			return false
		}
	}
	return true
}

type ortag []tag

func (o ortag) match(tags []string) bool {
	for _, t := range o {
		if t.match(tags) {
			return true
		}
	}
	return false
}

func split(s []byte, d byte) [][]byte {
	return bytes.Split(s, []byte{d})
}

func parseOr(tags []byte) tag {
	if len(tags) == 0 {
		return nil
	}
	ors := split(tags, ' ')
	if len(ors) == 1 {
		return parseAnd(ors[0])
	}
	var o ortag
	for _, t := range ors {
		o = append(o, parseAnd(t))
	}
	return o
}

func parseAnd(tags []byte) tag {
	//BUG(jmf): Tag parser does not handle invalid build tag sequence ,,
	ands := split(tags, ',')
	if len(ands) == 1 {
		return parse1(ands[0])
	}
	var a andtag
	for _, t := range ands {
		a = append(a, parse1(t))
	}
	return a
}

func parse1(tag []byte) tag {
	t := string(tag)
	if t[0] == '!' {
		return negtag(t[1:])
	}
	return atag(t)
}

func parseTags(r io.Reader) (tag, error) {
	lines, err := readUntilPackage(r)
	if err != nil {
		return nil, err
	}

	bt := extractBuildTags(lines)

	return parseOr(bt), nil
}
