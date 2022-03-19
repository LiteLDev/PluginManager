package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func newVmInstance() *otto.Otto {
	vm := otto.New()

	// jsFilesystem impl some simple functions for file Read, Write, etc.
	// it's basically a wrapper for Golang.OS
	type jsFilesystem struct {
		// Copy file from src to dst
		// jsRef: Copy(src: "./File1",dst: "./File2")
		// return file len if success, or error message if failed
		Copy func(call otto.FunctionCall) otto.Value

		// Delete file
		// jsRef: Delete(file: "./File1")
		// return null if success, or error message if failed
		Delete func(call otto.FunctionCall) otto.Value

		// Check if file exists
		// jsRef: Exists(file: "./File1")
		// return true if file exists, or false if not
		Exists func(call otto.FunctionCall) otto.Value

		// Create a new file with default permission
		// jsRef: Create(file: "./File1")
		// return null if success, or error message if failed
		Create func(call otto.FunctionCall) otto.Value

		// Make a directory
		// jsRef: Mkdir(dir: "./Dir1")
		// return null if success, or error message if failed
		Mkdir func(call otto.FunctionCall) otto.Value

		// Read file content
		// jsRef: Read(file: "./File1")
		// return file content if success, or error message if failed
		Read func(call otto.FunctionCall) otto.Value

		// Clear and write file content
		// jsRef: Write(file: "./File1", content: "Hello World")
		// return null if success, or error message if failed
		Write func(call otto.FunctionCall) otto.Value

		// Append content to file
		// jsRef: Append(file: "./File1", content: "Hello World")
		// return null if success, or error message if failed
		Append func(call otto.FunctionCall) otto.Value
	}

	type jsSystem struct {
		Cmd func(call otto.FunctionCall) otto.Value
	}

	sys := jsSystem{
		Cmd: func(call otto.FunctionCall) otto.Value {
			args := make([]string, len(call.ArgumentList)-1)
			for k, arg := range call.ArgumentList[1:] {
				args[k] = arg.String()
			}
			err := exec.Command(call.Argument(0).String(), args...).Run()
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			return otto.NullValue()
		},
	}

	fs := jsFilesystem{
		Copy: func(call otto.FunctionCall) otto.Value {
			//copy file wrapper
			count, err := func(src, dst string) (int64, error) {
				sourceFileStat, err := os.Stat(src)
				if err != nil {
					return 0, err
				}

				if !sourceFileStat.Mode().IsRegular() {
					return 0, fmt.Errorf("%s is not a regular file", src)
				}

				source, err := os.Open(src)
				if err != nil {
					return 0, err
				}
				defer source.Close()

				destination, err := os.Create(dst)
				if err != nil {
					return 0, err
				}
				defer destination.Close()
				nBytes, err := io.Copy(destination, source)
				return nBytes, err
			}(call.Argument(0).String(), call.Argument(1).String())

			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			ret, _ := vm.ToValue(count)
			return ret
		},
		Delete: func(call otto.FunctionCall) otto.Value {
			//delete file wrapper
			err := os.Remove(call.Argument(0).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			return otto.NullValue()
		},
		Exists: func(call otto.FunctionCall) otto.Value {
			//exists file wrapper
			_, err := os.Stat(call.Argument(0).String())
			if err != nil {
				ret, _ := vm.ToValue(false)
				return ret
			}
			ret, _ := vm.ToValue(true)
			return ret
		},
		Create: func(call otto.FunctionCall) otto.Value {
			//create file wrapper
			file, err := os.Create(call.Argument(0).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			defer file.Close()
			return otto.NullValue()
		},
		Mkdir: func(call otto.FunctionCall) otto.Value {
			//mkdir wrapper
			err := os.Mkdir(call.Argument(0).String(), 0777)
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			return otto.NullValue()
		},
		Read: func(call otto.FunctionCall) otto.Value {
			//read file wrapper
			file, err := os.Open(call.Argument(0).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			defer file.Close()
			bytes, err := ioutil.ReadAll(file)
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			ret, _ := vm.ToValue(string(bytes))
			return ret
		},
		Write: func(call otto.FunctionCall) otto.Value {
			//write file wrapper
			file, err := os.Create(call.Argument(0).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			defer file.Close()
			_, err = file.WriteString(call.Argument(1).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			return otto.NullValue()
		},
		Append: func(call otto.FunctionCall) otto.Value {
			//append file wrapper
			file, err := os.OpenFile(call.Argument(0).String(), os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			defer file.Close()
			_, err = file.WriteString(call.Argument(1).String())
			if err != nil {
				ret, _ := vm.ToValue(err.Error())
				return ret
			}
			return otto.NullValue()
		},
	}

	vm.Set("system", sys)
	vm.Set("filesystem", fs)

	return vm

}
