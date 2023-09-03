package envygo

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

type OsEnv struct {
	Create     func(string) (*os.File, error)
	Remove     func(string) error
	Path       string
	Properties interface{}
	umask      int
}

type UtilEnv struct {
	Log func(...interface{}) (int, error)
}

func TestEnvMockRandom(t *testing.T) {
	path := "file:///this/is/sample/path"
	var env = OsEnv{
		Create:     os.Create,
		Remove:     os.Remove,
		Path:       path,
		Properties: struct{ property string }{property: "sample property"},
		umask:      0x022,
	}

	if env.Path != path {
		t.Errorf("paths are not the same even before mocking!")
	}

	url := "http://host:port/this/is/sample/path"
	defer Mock(&env, &struct {
		Create func(string) (*os.File, error)
		Path   string
		umask  int
	}{
		func(s string) (*os.File, error) {
			return nil, errors.New("cant create a file")
		},
		url,
		0,
	})()

	if env.Path != url {
		t.Errorf("path should have been changed to url")
	}
}

func TestEnvMockSame(t *testing.T) {
	path := "file:///this/is/sample/path"
	var env = OsEnv{
		Create:     os.Create,
		Remove:     os.Remove,
		Path:       path,
		Properties: struct{ property string }{property: "sample property"},
		umask:      0x022,
	}

	if env.Path != path {
		t.Errorf("paths are not the same even before mocking!")
	}

	url := "http://host:port/this/is/sample/path"
	defer Mock(&env, &OsEnv{
		Create: func(s string) (*os.File, error) {
			return nil, errors.New("cant create a file")
		},
		Remove: nil,
		Path:   url,
		Properties: struct {
			blah int
		}{10},
	})()

	if env.Path != url {
		t.Errorf("path should have been changed to url")
	}
}

func TestMock(t *testing.T) {
	path := "file:///this/is/sample/path"
	var env = OsEnv{
		Create:     os.Create,
		Remove:     os.Remove,
		Path:       path,
		Properties: struct{ property string }{property: "sample property"},
		umask:      0x022,
	}

	Introduce(&env, &UtilEnv{})

	defer MockMany(
		&OsEnv{
			Create: func(s string) (*os.File, error) {
				return nil, errors.New("cant mock")
			},
		},
		&UtilEnv{
			Log: fmt.Println,
		})()

	if _, err := env.Create("hello"); err == nil {
		t.Errorf("error was expected")
	}
}

func TestUnexportedMocking(t *testing.T) {
	var env = OsEnv{}

	const perm = 0o022
	defer Mock(&env, &struct {
		Path  string
		umask int
	}{"Hello", perm})()

	if perm != env.umask {
		t.Fail()
	}
}

func TestMockField(t *testing.T) {
	var env = OsEnv{}

	const perm = 0o022
	defer MockField(&env, "umask", perm)()

	if perm != env.umask {
		t.Fail()
	}
}

func TestMockFields(t *testing.T) {
	var env = OsEnv{}

	const perm = 0o022
	const path = "path to my home"
	defer MockFields(&env, Fields{"umask": perm, "Path": path})()

	if perm != env.umask || path != env.Path {
		t.Fail()
	}
}

func helperResetMockFields(env interface{}, tester func()) {
	const perm = 0o022
	const path = "path to my home"
	defer MockFields(env, Fields{"umask": perm, "Path": path})()
	tester()
}

func TestResetMockFields(t *testing.T) {
	var env = OsEnv{}
	perm := env.umask
	path := env.Path
	var called bool
	ptr := &called
	helperResetMockFields(&env, func() {
		*ptr = true
		if perm == env.umask || path == env.Path {
			t.Fail()
		}
	})

	if !called || perm != env.umask || path != env.Path {
		t.Fail()
	}
}
