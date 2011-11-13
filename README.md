# GoRun - an easier way to build and run go programs.

I don't always write go code, but when I do I loathe using Makefiles. Instead of edit-make-run, I wanted to simply run the go source files directly. gorun makes that possible.

## Getting Started
    // hello.go
    package main
    import "fmt"
    func main() {
      fmt.Println("hello go")
    }

-

    $ gorun hello.go
    hello go

## Options

__--goroot=directory__ the path to your _GOROOT_.

__--build-dir=directory__ a path to which build artifacts are written. By default, the artifacts will be written to the directory __out__.

__--output=file__ create the executable file at _file_.

__--alias=package:import__ allows you to build a package to a particular import path. A good example is the mgo library internally imports "launchpad.net/mgo" and, of course, its package name is _mgo_. To allow the use of import "launchpad.net/mgo", simply provide the flag --alias=mgo:launchpad.net/mgo.

## Common Example

__main.go__

    package main
    import (
      "fmt"
      "mylib"
    )
    func main() {
      fmt.Println(mylib.Name())
    }

__mylib.go__

    package mylib
    func Name() string {
      return "MyLib"
    }

-

    $ gorun *.go
    MyLib
