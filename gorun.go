package main

import (
  "errors"
  "fmt"
  "go/parser"
  "go/token"
  "path/filepath"
  "os"
  "strings"
)

// TODO:
// 1 - Comment some of this code.
// 2 - Add checks of unnecessary builds.
// 3 - Add meta-data file format.

func defaultGoRoot() string {
  env := os.Getenv("GOROOT")
  if env != "" {
    return env
  }

  return fmt.Sprintf("%s/src/go", os.Getenv("HOME"))
}

type lib struct {
  Name string
  Alias string
  Files []string
  Imports map[string]bool
}

func newLib(name, alias string) *lib {
  return &lib{name, alias, make([]string, 0), make(map[string]bool)}
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

func createBuild(files []string, aliases map[string]string) ([]*lib, error) {
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
    alias, ok := aliases[pkg]
    if !ok {
      alias = pkg
    }

    lib := libs[alias]
    if lib == nil {
      lib = newLib(pkg, alias)
      libs[alias] = lib
    }

    lib.addFile(files[i])
    imports := ast.Imports
    for _, path := range imports {
      v := path.Path.Value
      lib.addImport(v[1 : len(v) - 1])
    }
  }

  // Next, flatten that map into a build list.
  return flattenBuild(libs["main"], libs, map[string]bool{}, []*lib{}), nil
}

func splitArgs() ([]string, []string) {
  args := os.Args
  for i := 0; i < len(args); i++ {
    if args[i] == "--" {
      return args[1:i], args[i + 1:]
    }
  }
  return args[1:], args[len(args):]
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
  out6 := filepath.Join(buildDir, fmt.Sprintf("%s.6", lib.Alias))

  dirname, _ := filepath.Split(out6)
  os.MkdirAll(dirname, 0700)

  args := []string{fmt.Sprintf("-I%s", buildDir), "-o", out6}
  if !call(filepath.Join(goroot, "bin/6g"), append(args, lib.Files...)...) {
    return false
  }

  outa := filepath.Join(buildDir, fmt.Sprintf("%s.a", lib.Alias))
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

type options struct {
  Files []string
  Output string
  GoRoot string
  BuildDir string
  Aliases map[string]string
  ProgramArgs []string
}

const (
  outputFlagPrefix = "--output="
  buildDirFlagPrefix = "--build-dir="
  gorootFlagPrefix = "--goroot="
  aliasFlagPrefix = "--alias="
)

func parseArgs() (*options, error) {
  args, pargs := splitArgs()

  output := ""
  buildDir := "out"
  goroot := defaultGoRoot()
  aliases := map[string]string{}

  i := 0
  for ; i < len(args); i++ {
    arg := args[i]

    // If this doesn't look like a flag, proceed.
    if arg[0] != '-' && arg[1] != '-' {
      break
    }

    // --output
    if strings.HasPrefix(arg, outputFlagPrefix) {
      output = arg[len(outputFlagPrefix):]
      continue
    }

    // --build-dir
    if strings.HasPrefix(arg, buildDirFlagPrefix) {
      buildDir = arg[len(buildDirFlagPrefix):]
      continue
    }

    // --goroot
    if strings.HasPrefix(arg, gorootFlagPrefix) {
      goroot = arg[len(gorootFlagPrefix):]
      continue
    }

    // --alias
    if strings.HasPrefix(arg, aliasFlagPrefix) {
      vals := strings.SplitN(arg[len(aliasFlagPrefix):], ":", 2)
      if len(vals) != 2 {
        return nil, errors.New(fmt.Sprintf("Invalid flag: %s", arg))
      }
      aliases[vals[0]] = vals[1]
      continue
    }

    return nil, errors.New(fmt.Sprintf("Invalid flag: %s", arg))
  }

  return &options{
    args[i : len(args)],
    output,
    goroot,
    buildDir,
    aliases,
    pargs}, nil
}

func main() {
  options, err := parseArgs()
  if err != nil {
    // todo: fix this
    panic(err)
  }

  libs, err := createBuild(options.Files, options.Aliases)
  if err != nil {
    panic(err)
  }

  // Ensure that buildDir exits.
  os.MkdirAll(options.BuildDir, 0755)

  // Build all lib dependencies.
  for _, lib := range libs[:len(libs) - 1] {
    if !buildLib(options.GoRoot, options.BuildDir, lib) {
      os.Exit(1)
    }
  }

  // Build the main binary.
  main := libs[len(libs) - 1]
  dest := outputTo(options.Output, options.BuildDir, main)
  if !buildApp(options.GoRoot, options.BuildDir, main, dest) {
    os.Exit(1)
  }

  // Execute the main binary.
  if !call(dest, options.ProgramArgs...) {
    os.Exit(1)
  }
}
