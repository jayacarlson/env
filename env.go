package env

import (
	"encoding/binary"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

/* ========================================================================= //
	NOTE:  The startup order for GO is:
		Any variable declaration
		Any init() func
		Then MAIN

	So, we use variable declaration to make sure the env is read in before
	other package 'init' functions, which could use this env package.

	ReadEnvVars will handle strings, ints & []strings -- see envSep below

	An outside package can call ReadEnvVars to retrieve any environment vars
	specific for it:

		var myEnvVars struct {
			Foo    string // names must be capitalized for this package to read them
			Bar    int
			foobar data{} // non-cap names are private and aren't touched
		}

		func init() {
			env.ReadEnvVars(&myEnvVars)
		}
// ------------------------------------------------------------------------- */

var (
	envSep = getEnv() // doing this gets the environment vars before any init() function(s) are called
	//                   also gives what to split any string slices with, ':' for linux, ';' for windows

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

// return if system is little endian
func ImLittleEndian() bool {
	et := 1
	return *(*byte)(unsafe.Pointer(&et)) == 1
}

// return if system is little endian
func ImBigEndian() bool {
	return !ImLittleEndian()
}

// return proper system encoding
func MyEncoding() binary.ByteOrder {
	if ImLittleEndian() {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

// return non native encoding
func NotMyEncoding() binary.ByteOrder {
	if ImBigEndian() {
		return binary.LittleEndian
	}
	return binary.BigEndian
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
func getEnv() string {
	sep := ":"
	ReadEnvVars(&env)

	// validate we have some values
	if env.Host == "" {
		env.Host = runtime.GOOS
	}
	if env.User == "" {
		// try Windows 'USERNAME'
		getEnvVal("USERNAME", reflect.ValueOf(&env).Elem().FieldByName("User"))
	}
	if env.IsWindows() {
		sep = ";"
	}

	return sep
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
