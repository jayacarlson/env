package env

import (
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

/* ========================================================================= //
	NOTE:  The startup order for GO is:
		Any variable declaration
		Any init() func
		Then MAIN

	So, we use variable declaration to make sure the env is read in before
	other packages 'init' functions which could use the env package.

	ReadEnvVars will handle strings, ints & []strings -- see envSep below

	An outside package can call ReadEnvVars to retrieve any environment vars
	specific for it:

	var myEnvVars struct {
		Foo string // names must be capitalized for this package to read them
		Bar int
	}

	func init() {
		env.ReadEnvVars(&myEnvVars)
	}
// ------------------------------------------------------------------------- */

const envSep = ":" // what to split any string slices with

var (
	envSet = getEnv() // doing this gets the environment vars before any init() function(s) are called

	env struct {
		Host string // host name (read on linux, assigned on wondows)
		User string // user name (read on linux, re-read from username on windows)
	}
)

// return current HOST system: 'linux' | 'windows'
func Host() string {
	return env.Host
}

// return current USER name
func User() string {
	return env.User
}

// simple boolean if system is 'linux'
func IsLinux() bool {
	return env.Host == "linux"
}

// simple boolean if system is 'windows'
func IsWindows() bool {
	return env.Host == "windows"
}

// read the env vars and try matching them into any structure passed
func ReadEnvVars(i interface{}) {
	v := reflect.ValueOf(i).Elem()
	t := v.Type()

	// Override default values with environment variables
	for i := 0; i < v.NumField(); i++ {
		getEnvVal(strings.ToUpper(t.Field(i).Name), v.Field(i))
	}
}

// getEnv -- run as variable assignment to be assured it is run before all 'init' methods; some which may call into here
func getEnv() bool {
	ReadEnvVars(&env)

	// validate we have some values
	if env.Host == "" {
		env.Host = runtime.GOOS
	}
	if env.User == "" {
		// try Windows 'USERNAME'
		getEnvVal("USERNAME", reflect.ValueOf(&env).Elem().FieldByName("User"))
	}

	return true
}

// read in env vars for element
func getEnvVal(envname string, field reflect.Value) {
	envVal := os.Getenv(envname)

	if len(envVal) > 0 {
		switch field.Kind() {
		case reflect.String:
			field.Set(reflect.ValueOf(envVal))
		case reflect.Int:
			v, err := strconv.Atoi(envVal)
			if err != nil {
				panic("ReadEnvVars: Illegal atoi conversion")
			}
			field.Set(reflect.ValueOf(v))
		case reflect.Slice:
			switch field.Type() {
			case reflect.TypeOf([]string(nil)):
				v := strings.Split(envVal, envSep)
				field.Set(reflect.ValueOf(v))
			default:
				panic("ReadEnvVars: Unexpected type")
			}
		default:
			panic("ReadEnvVars: Unexpected kind")
		}
	}
}
