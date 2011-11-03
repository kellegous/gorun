package main

import (
  "flag"
  "fmt"
  "go/parser"
  "go/token"
  "os"
)

var output = flag.String("output", "main", "")
var goroot = flag.String("goroot", "/Users/knorton/src/go", "")
var buildDir = flag.String("build-dir", "out", "")

type lib struct {
  Name string
  Files []string
  Imports map[string]bool
}

func newLib(name string) *lib {
  return &lib{name, make([]string, 0), make(map[string]bool)}
}

func (l *lib) addImport(path string) {
  l.Imports[path] = true
}

func (l *lib) addFile(name string) {
  l.Files = append(l.Files, name)
}

func createBuild(files []string) ([]*lib, []string, os.Error) {
  libs := make(map[string]*lib)
  // First build a map of libs.
  for i := 0; i < len(files); i++ {
    ast, err := parser.ParseFile(token.NewFileSet(),
      files[i],
      nil,
      parser.PackageClauseOnly & parser.ImportsOnly)
    if err != nil {
      return nil, nil, err
    }

    pkg := ast.Name.String()
    if pkg == "main" {
    }

    lib := libs[pkg]
    if lib == nil {
      lib = newLib(pkg)
      libs[pkg] = lib
    }

    lib.addFile(files[i])
    imports := ast.Imports
    for j := 0; j < len(imports); j++ {
      path := imports[j].Path.Value
      lib.addImport(path[1 : len(path) - 1])
    }
  }

  for _, v := range(libs) {
    fmt.Println(v)
  }
  // Next, flatten that map into a build list.
  return nil, nil, nil
}

func splitArgs() ([]string, []string) {
  files := make([]string, 0)
  pargs := make([]string, 0)
  i := 0
  for ; i < flag.NArg(); i++ {
    arg := flag.Arg(i)
    if flag.Arg(i) == "--" {
      i++
      break
    }
    files = append(files, arg)
  }

  for ; i < flag.NArg(); i++ {
    pargs = append(pargs, flag.Arg(i))
  }

  return files, pargs
}

func main() {
  flag.Parse()
  fmt.Printf("output:     %s\n", *output)
  fmt.Printf("groot:      %s\n", *goroot)
  fmt.Printf("build-dir:  %s\n", *buildDir)

  files, _ := splitArgs()
  createBuild(files)
}
