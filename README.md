# envygo
Environment aware mocking library for golang
```golang
import env github.com/celeral/envygo
```
## motivation
So much has been made out of mocking while testing. There are elaborate framewors where a Golang coder is expected to create the interfaces and then generate code from these interfaces that can be used to verify certain behavior. Many of these frameworks are wildly accepted, probably because these framework capitalized on a certain taste the developers developed and came to expect from mocking libraries prior to advent of Golang. Yet article after article, especially when we talk about missing polymorphism in Golang, we talk about how the function of Go function should be modified by passing it a function as an argument.

Really... Mocking should not be so invasive or convoluted. It need not have its own chapter... or may be having it does help to stress the importance of writing meaningful unit tests; May be a chapter is indeed needed... to save the developers from the trap of writing integration tests disguised as unit tests. I digress.

So without further adieu, how about some code. 

### src
```golang
func ReadConfiguration(relativePath string) Config, error {
    absolutePath := "/etc/" + relativePath

    bytes, err := Os.ReadFile(absolutePath)
    if  err == nil {
       return ParseConfig(bytes)
    }
    return nil, err
}
```

### test
```golang
func TestReadConfigurationCaseEmptyFile(t *testing.T) {
	defer MockField(Os, "ReadFile", func(name string) ([]byte, error) { return []byte{}, nil })()

	if config, err := ReadConfiguration("this/path/does/not/exist"); err != nil {
		  if !config.IsEmpty() {
			    t.Errorf("Config is not empty %v", config)
		  }
	} else {
		  t.Errorf("Failed %v", err)
	}
}
```

How is this possible, I hear some of you ask. Only "some" because others probably noticed it's `Os.ReadFile` and not `os.ReadFile`

### src again
```golang
import env github.com/celeral/envygo

var Mock = env.Mock
var Unmock = env.Unmock
var MockField = env.MockField
 
var Os = &struct {
        Create   func(name string) (*os.File, error)
	ReadFile func(name string) ([]byte, error)
}{
  // for now these are the only 2 functions I would want to override
  ReadFile: os.ReadFile,
  Create:   os.Create,
}
```

## idiom

Forget about mocking frameworks, passing functions as arguments and spending hours thinking about modification of code structure and then actually modifying it and then doing it again.

Instead the functions that you want to change behavior of during testing or even based on environment (yup, that's where env-y go comes from), define a few global variables of type `struct` in your code and invoke your code via fields of these structures - only for the code which you intend to mock. The examples in code above are `os.ReadFile` and `os.Create`. To standardize in a minimally invasive way - I decided to name my global variable `Os`. So my code now calls `Os.ReadFile` instead of `os.ReadFile`

## examples

my favorite

```golang
// package.go
type ConstantsEnv struct {
    baseDirectory string
    ConfigurationPath string
}

var Constants = &ConstantsEnv{
    baseDirectory:     "/opt/data/mypackage",
    ConfigurationPath: "etc/package.conf"
}

// source.go
func doSomething() {
    configurationFile := Constants.baseDirectory + "/" + Constants.ConfigurationPath

     // code to really do something with configurationFile
}


// source_test.go
func TestDoSomething(t *testing.T)
{
   defer Unmock(Mock(Constants, &ConstantsEnv{ baseDirectory: "testdata" })) // specify ConfigurationPath as well if you dont like original one

   doSomething()

   // code to verify that something was really done with our test configurationFile by doSomething
}
```

## other features

The footprint of the envygo is tiny. You will easily figure out what it has to offer by looking at source. But one cryptic thing is support for parallelism while mocking. When running tests in parallel, if you dont want one test's environment modeling (mocking) clobbering that of another then you would want the later one to wait until the former is done. To achieve it you can do one of the following.

```golang
type MyEnv struct {
   mutex           sync.Mutex `env:"mutex"` // special tag to identify mutex for var of type MyEnv
   doSomething     func()
   interestingPath string
}
```
or
```golang
type MyEnv struct {
   mutex           *sync.Mutex `env:"mutex"` // or it can be a pointer instead of struct
   doSomething     func()
   interestingPath string
}
```
or
```golang
type MyEnv struct {
   locker          func(*MyEnv, bool) `env:"mutex"` // or do something more fun using the locker function
   doSomething     func()
   interestingPath string
}
```

## feedback

what do you think? stars, issues, emails - I am ears!



