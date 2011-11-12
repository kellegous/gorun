package main
import (
  "fmt"
  "os"
)
func main() {
  if os.Args[1] == "--arg0" && os.Args[2] == "-arg1" && os.Args[3] == "2" {
    fmt.Println("ok")
  } else {
    panic("failed")
  }
}
