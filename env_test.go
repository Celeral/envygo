package envygo

import (
	"errors"
	"fmt"
	"os"
	"sync"
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

type mutexEnv struct {
	mutex sync.Mutex `env:"mutex"`
	Name  string
}

func TestMutex(t *testing.T) {
	var env = &mutexEnv{Name: "original"}
	go func() { defer Unmock(MockField(env, "Name", "mocked 1")) }()
	go func() { defer Unmock(MockField(env, "Name", "mocked 2")) }()
}

type mutexPtrEnv struct {
	mutex *sync.Mutex `env:"mutex"`
	Name  string
}

func TestMutexPtr(t *testing.T) {
	var env = &mutexPtrEnv{mutex: &sync.Mutex{}, Name: "original"}

	var latch sync.WaitGroup
	latch.Add(2)

	go func() {
		defer Unmock(MockField(env, "Name", "mocked1"))
		if env.Name != "mocked1" {
			t.Fail()
		}
		latch.Done()
	}()

	go func() {
		defer Unmock(MockField(env, "Name", "mocked2"))
		if env.Name != "mocked2" {
			t.Fail()
		}
		latch.Done()
	}()

	latch.Wait()
}

type mutexFuncEnv struct {
	mutex Locker `env:"mutex"`
	Name  string
}

func TestMutexFunc(t *testing.T) {
	var lockError bool
	mutex := sync.Mutex{}
	var env = &mutexFuncEnv{mutex: func(old interface{}, lockUnlock bool) {
		if lockUnlock {
			if !mutex.TryLock() {
				lockError = true
			}
		} else if !lockError {
			mutex.Unlock()
		}
	}, Name: "original"}
	defer Unmock(MockField(env, "Name", "mocked 1"))
	defer Unmock(MockField(env, "Name", "mocked 2"))
}
