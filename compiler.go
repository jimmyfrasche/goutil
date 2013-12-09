package goutil

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//Gostrap creates a temporary environment to run the go(1) tool in.
//
//It embeds an exec.Cmd configured to run the go tool in this environment.
type Gostrap struct {
	root, old string
	*exec.Cmd
}

//NewGostrap creates a new Gostrap.
//
//The embedded exec.Cmd still needs the tool (ie, run, install, etc) and its
//arguments appended to its Args list. It defaults to using os.Stdout
//and os.Stderr for the Stdout and Stderr of the Cmd, respectively.
//
//You should investigate Run and Install first, as they are likely what you
//need, and if not their source, and WithGostrap's, show how Gostrap is used.
func NewGostrap() (*Gostrap, error) {
	t, err := ioutil.TempDir("", "lib-goutil")
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Join(t, "src", "src"), 0777); err != nil {
		os.RemoveAll(t)
		return nil, err
	}
	gop, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}
	env := os.Environ()
	for i, e := range env {
		if strings.HasPrefix(e, "GOPATH=") {
			env[i] = e + ":" + t
			break
		}
	}
	tmp := &Gostrap{
		root: t,
		Cmd: &exec.Cmd{
			Path:   gop,
			Env:    env,
			Args:   []string{"go"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
	}
	return tmp, nil
}

//Destroy the Gostrap and everything in it and everything it loves.
func (t *Gostrap) Destroy() {
	//ignore error, we're just being nice by calling this
	t.Popdir()

	//don't want to do anything vicious if this the zero value
	if t.root != "" {
		//There is nothing to do if this fails and no semantic consequence,
		//so it may do so in silence.
		os.RemoveAll(t.root)
	}
}

//Chdir makes the temp location the current working directory.
//It saves the current directory so this can be undone with Popdir.
func (t *Gostrap) Chdir() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err = os.Chdir(t.JoinPath()); err != nil {
		return err
	}
	t.old = cwd
	return nil
}

//Popdir returns the working directory to what it was before Chdir was called.
func (t *Gostrap) Popdir() error {
	if t.old == "" {
		return errors.New("Never changed directory")
	}
	if err := os.Chdir(t.old); err != nil {
		return err
	}
	t.old = ""
	return nil
}

//JoinPath joins path to the temp directory.
func (t *Gostrap) JoinPath(path ...string) string {
	return filepath.Join(t.root, "src", "src", filepath.Join(path...))
}

//AddFile creates a file in the temp location.
func (t *Gostrap) AddFile(name string, contents []byte) error {
	return ioutil.WriteFile(t.JoinPath(name), contents, 0666)
}

//WithGostrap creates a new Gostrap, calls Chdir, and defers
//calls to PopDir and Destroy, in the appropriate order, before invoking f.
func WithGostrap(f func(*Gostrap) error) error {
	t, err := NewGostrap()
	if err != nil {
		return err
	}
	defer t.Destroy()

	err = t.Chdir()
	if err != nil {
		return err
	}
	defer t.Popdir()

	return f(t)
}

func tagargs(tags []string) (out []string) {
	if len(tags) > 0 {
		//XXX what the hell is the format for multiple tags?
		out = append(out, "-tags", strings.Join(tags, " "))
	}
	return
}

//Run calls go run on a single input file, with optional build tags,
//in a Gostrap that's created on the fly and destroyed after.
func Run(file []byte, tags ...string) error {
	return WithGostrap(func(t *Gostrap) error {
		err := t.AddFile("main.go", file)
		if err != nil {
			return err
		}

		t.Args = append(t.Args, "run")
		t.Args = append(t.Args, tagargs(tags)...)
		t.Args = append(t.Args, "main.go")

		return t.Run()
	})
}

//Install calls go install on a single input file, with optional build tags,
//in a Gostrap that's created on the fly and destroyed after.
func Install(name, location string, file []byte, tags ...string) error {
	if location == "" || location == "." {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		location = pwd
	}
	return WithGostrap(func(t *Gostrap) error {
		err := t.AddFile("main.go", file)
		if err != nil {
			return err
		}

		t.Args = append(t.Args, "build")
		t.Args = append(t.Args, tagargs(tags)...)
		t.Args = append(t.Args, "-o", name, "main.go")

		err = t.Run()
		if err != nil {
			return err
		}
		//need to copy the file
		return os.Rename(name, filepath.Join(location, name))
	})
}
