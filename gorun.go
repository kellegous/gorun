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

func (l *lib) isMain() bool {
  return l.Name == "main"
}

func flattenBuild(target *lib, libs map[string]*lib, seen map[string]bool, build []*lib) []*lib {
  for name, _ := range target.Imports {
    lib := libs[name]
    if lib == nil {
      continue
    }
    if _, ok := seen[lib.Name]; ok {
      continue
    }

    for _, dep := range flattenBuild(lib, libs, seen, build) {
      build = append(build, dep)
    }
  }

  seen[target.Name] = true
  return append(build, target)
}

func createBuild(files []string) ([]*lib, os.Error) {
  libs := make(map[string]*lib)

  // First build a map of libs.
  for i := 0; i < len(files); i++ {
    ast, err := parser.ParseFile(token.NewFileSet(),
      files[i],
      nil,
      parser.PackageClauseOnly & parser.ImportsOnly)
    if err != nil {
      return nil, err
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

  // Next, flatten that map into a build list.
  return flattenBuild(libs["main"], libs, make(map[string]bool), make([]*lib, 0)), nil
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

func buildLib(goroot, buildDir string, lib *lib) bool {
  fmt.Printf("Build Lib %s\n", lib.Name)
  return true
}

func buildApp(gooroot, buildDir string, lib *lib) bool {
  fmt.Printf("Build App %s\n", lib.Name)
  return true
}

func main() {
  flag.Parse()
  fmt.Printf("output:     %s\n", *output)
  fmt.Printf("groot:      %s\n", *goroot)
  fmt.Printf("build-dir:  %s\n", *buildDir)

  files, _ := splitArgs()
  libs, err := createBuild(files)
  if err != nil {
    panic(err)
  }

  for _, lib := range libs[:len(libs) - 1] {
    buildLib(*goroot, *output, lib)
  }
  buildApp(*goroot, *output, libs[len(libs) - 1])
}
