package main

import (
  "flag"
  "fmt"
  "go/parser"
  "go/token"
  "path/filepath"
  "os"
)

// TODO:
// 1 - Comment some of this code.
// 2 - Add checks of unnecessary builds.
// 3 - Add meta-data file format.
var output = flag.String("output", "", "")
var goroot = flag.String("goroot", defaultGoRoot(), "")
var buildDir = flag.String("build-dir", "out", "")

func defaultGoRoot() string {
  env := os.Getenv("GOROOT")
  if env != "" {
    return env
  }

  return fmt.Sprintf("%s/src/go", os.Getenv("HOME"))
}

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
    build = append(build, flattenBuild(lib, libs, seen, build)...)
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
  return flattenBuild(libs["main"], libs, map[string]bool{}, []*lib{}), nil
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

func call(command string, args ...string) bool {
  c := append([]string{command}, args...)
  p, err := os.StartProcess(
      command,
      c,
      &os.ProcAttr{
        "",
        os.Environ(),
        []*os.File{nil, os.Stdout, os.Stderr},
        nil})
  if err != nil {
    return false
  }
  s, err := p.Wait(0)
  if err != nil {
    return false
  }
  return s.WaitStatus.ExitStatus() == 0
}

func buildLib(goroot, buildDir string, lib *lib) bool {
  out6 := filepath.Join(buildDir, fmt.Sprintf("%s.6", lib.Name))
  args := []string{fmt.Sprintf("-I%s", buildDir), "-o", out6}
  if !call(filepath.Join(goroot, "bin/6g"), append(args, lib.Files...)...) {
    return false
  }

  outa := filepath.Join(buildDir, fmt.Sprintf("%s.a", lib.Name))
  if !call(filepath.Join(goroot, "bin/gopack"), "grc", outa, out6) {
    return false
  }

  return true
}

func buildApp(goroot, buildDir string, lib *lib, output string) bool {
  out6 := filepath.Join(buildDir, fmt.Sprintf("%s.6", lib.Name))
  args := []string{fmt.Sprintf("-I%s", buildDir), "-o", out6}
  if !call(filepath.Join(goroot, "bin/6g"),
      append(args, lib.Files...)...) {
    return false
  }

  if !call(filepath.Join(goroot, "bin/6l"),
      fmt.Sprintf("-L%s", buildDir),
      "-o",
      output,
      out6) {
    return false
  }

  return true
}

func outputTo(flag string, buildDir string, lib *lib) string {
  if flag != "" {
    return flag
  }
  return filepath.Join(buildDir, lib.Name)
}

func main() {
  flag.Parse()

  files, args := splitArgs()
  libs, err := createBuild(files)
  if err != nil {
    panic(err)
  }

  // Ensure that buildDir exits.
  os.MkdirAll(*buildDir, 0755)

  // Build all lib dependencies.
  for _, lib := range libs[:len(libs) - 1] {
    if !buildLib(*goroot, *buildDir, lib) {
      os.Exit(1)
    }
  }

  // Build the main binary.
  main := libs[len(libs) - 1]
  dest := outputTo(*output, *buildDir, main)
  if !buildApp(*goroot, *buildDir, main, dest) {
    os.Exit(1)
  }

  // Execute the main binary.
  if !call(dest, args...) {
    os.Exit(1)
  }
}
