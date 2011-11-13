### Overview

I don't always write go code, but when I do I loathe using Makefiles. Instead of edit-make-run, I wanted to simply run the go source files directly. gorun makes that possible.

For example:

  hello.go
    package main
    import "fmt"
    func main() {
      fmt.Println("Hello World")
    }

    $ gorun hello.go
    Hello World
