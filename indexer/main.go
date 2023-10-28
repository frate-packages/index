package main

import (
	"bufio"
	"encoding/json"
	"fmt"
  "os/exec"
	"os"
	"path/filepath"
	"strings"
  "regexp"
)

type PackageInfo struct{
  Name string `json:"name"`
  Versions []string `json:"versions"`
  Git string `json:"git"`
  Description string `json:"description"`
}


func validateVersionName(version string) bool{

  if(regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(version)){
    return true;
  }else if(regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version)){
    return true;
  }else if(version == "master"){
    return true;
  }else if(version == "latest"){
    return true;
  }else if(version == "stable"){
    return true;
  }else if(version == "main"){
    return true;
  }else{
    return false;
  }
}

func parseRemoteLsTags(output string) []string{
  lines := strings.Split(output, "\n")
  tags := []string{}

  for _, line := range lines{
    if(strings.Contains(line, "refs/tags")){
      tag := strings.Split(line, "\t")[1] 
      tag = strings.Replace(tag, "refs/tags/", "", -1)
      if(validateVersionName(tag)){
        tags = append(tags, tag)
      }
    }
    if(strings.Contains(line, "refs/heads")){
      tag := strings.Split(line, "\t")[1] 
      tag = strings.Replace(tag, "refs/heads/", "", -1)
      if(validateVersionName(tag)){
        tags = append(tags, tag)
      }
    }
  }
  return tags
}

func validGitLink(link string) bool{
  if(strings.Contains(link, "https://github")){
    return true;
  }else if(strings.Contains(link, "https://gitlab")){
    return true;
  }else if(strings.Contains(link, "https://bitbucket")){
    return true;
  }else if(strings.Contains(link, "svn")){
    return true;
  }else{
    return false;
  }
}


func main(){
  


  var packageIndex []PackageInfo;

  err := filepath.Walk("../index", func(path string, _ os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if(strings.Contains(path, "info.json")){
      file, err := os.Open(path)
      scanner := bufio.NewScanner(file)
      lines := ""

      if(err != nil){
        fmt.Println(err)
      }

      for scanner.Scan() {
        line := scanner.Text()
        lines += line
      }

      packageInfo := PackageInfo{}
      json.Unmarshal([]byte(lines), &packageInfo)

      if(validGitLink(packageInfo.Git)){

        if(len(packageInfo.Versions) < 1){

          if(err != nil){
            fmt.Println(err);
            return nil;
          }
          fmt.Println("Getting versions for " + packageInfo.Name)
          if(packageInfo.Name != "cub" && packageInfo.Name != "libliftoff" && packageInfo.Name != "libxpm"){
            cmd := exec.Command("git", "ls-remote", packageInfo.Git)


            out, err := cmd.CombinedOutput()

            if(err != nil){
              fmt.Println("Could not get versions for " + packageInfo.Name)
              return nil;
            }else{

              tags := parseRemoteLsTags(string(out))

              packageInfo.Versions = tags


              packageIndex = append(packageIndex, packageInfo)

            }
          }
        }else{
          packageIndex = append(packageIndex, packageInfo)
        }
      }
    }
    return nil
  })

  if err != nil {
    fmt.Println(err)
  }
  packageIndexJson, err := json.Marshal(packageIndex)

  if(err != nil){
    fmt.Println(err)
  }

  file, err := os.Create("../dist/index.json")
  
  fmt.Println("Writing index.json")


  if(err != nil){
    fmt.Println(err)
  }

  defer file.Close()


  file.WriteString(string(packageIndexJson));
}



