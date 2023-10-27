package main

import (
  "bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)



func main(){

  newIndex := "["

  err := filepath.Walk("./index", func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if(strings.Contains(path, "info.json")){
      file, err := os.Open(path)
      scanner := bufio.NewScanner(file)

      if(err != nil){
        fmt.Println(err)
      }

      for scanner.Scan() {
        line := scanner.Text()
        newIndex += line
      }

      newIndex += "\n,"
    }
    return nil
  })

  if err != nil {
    fmt.Println(err)
  }
  newIndex = newIndex[:len(newIndex)-1]
  newIndex += "]"
  file, err := os.Create("./dist/index.json")

  file.WriteString(newIndex)

  if err != nil {
    fmt.Println(err)
  }

  defer file.Close()
}
