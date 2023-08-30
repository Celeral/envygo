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
	defer MockOne(&env, struct {
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
	defer MockOne(&env, OsEnv{
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
