package main

import (

	"bufio"
	"fmt"
  "os/exec"
	"os"
	"path/filepath"
	"strings"
  "io"
  "regexp"
  "net/http"
  "encoding/json"
  "errors"

)
type PackageInfo struct{
  Name string `json:"name"`
  Versions []string `json:"versions"`
  Git string `json:"git"`
  GitShort string `json:"git_short"`
  GitPrefixed string `json:"git_prefixed"` 
  Stars int `json:"stars"`
  Forks int `json:"forks"`
  OpenIssues int `json:"open_issues"`
  Maintainers []string `json:"maintainers"`
  Watchers int `json:"watchers"`
  Description string `json:"description"`
  GitDescription string `json:"git_description"`
  Target string `json:"target_link"`
  License string `json:"license"`
  Language string `json:"language"`
  Owner string `json:"owner"`
  OwnerType string `json:"owner_type"`
}


func validateVersionName(version string) bool{
  //BECAUSE YOU FUCKERS CAN'T DECIDE ON YOUR FUCKING PACKAGE VERSIONING SCHEME
  //WE DID THIS SHIT ARE YOU HAPPY?!!!!

  //Typical versions like v1.2.3
  if(regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(version)){

    return true;
    //Typical versions like 1.2.3
  }else if(regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(version)){

    return true;
    //Typical versions like 1.2
  }else if(regexp.MustCompile(`^\d+\.\d+`).MatchString(version)){

    return true;
    //Typical versions like v1.2
  }else if(regexp.MustCompile(`v\d+.\d+`).MatchString(version)){

    return true;
    //Curl style versions like word-1_2_3
  }else if(regexp.MustCompile(`[A-Za-z]+-\d+_\d+_\d+`).MatchString(version)){

    return true;
    //word-1.2.3
  }else if(regexp.MustCompile(`[A-Za-z]+-\d+\.\d+\.\d+`).MatchString(version)){

    return true;
    //word_1.2.3
  }else if(regexp.MustCompile(`[A-Za-z]+_\d+\.\d+\.\d+`).MatchString(version)){

    return true;
    //word_1.2
  }else if(regexp.MustCompile(`[A-Za-z]+_\d+.\d+`).MatchString(version)){

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
  fmt.Printf("Found %d tags ",len(tags))
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

func packageOnlyHasMainTag(packageInfo PackageInfo) bool{
  if(len(packageInfo.Versions) == 1 && ( packageInfo.Versions[0] == "main" || packageInfo.Versions[0] == "master")){

    return true;

  }else{

    return false;

  }
}

func isGithubRepo(repo string) bool{
  if(strings.Contains(repo, "github.com")){
    return true;
  }else{
    return false;
  }
}
func isGitlabRepo(repo string) bool{

  if(strings.Contains(repo, "gitlab.com")){

    return true;

  }else{

    return false;

  }

}
func shortenGitLink(link string) string{
  if(isGithubRepo(link)){

    link = strings.Replace(link, "https://github.com/", "", -1)
    link = strings.Replace(link, ".git", "", -1)

    return link

  }else if(isGitlabRepo(link)){

    link = strings.Replace(link, "https://gitlab.com/", "", -1)
    link = strings.Replace(link, ".git", "", -1)

    return link

  }else{

    return link

  }
}
func isShortened(git string ) bool{
  expected_expr := "[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"
  return regexp.MustCompile(expected_expr).MatchString(git);
}
func makeAuthorizedRequest(url string, token string) string{
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Add("Authorization", "token " + token)
  req.Header.Add("Accept", "application/vnd.github.v3+json")
  req.Header.Add("User-Agent", "curl/7.64.1")
  resp, _ := http.DefaultClient.Do(req)
  defer resp.Body.Close()
  body, _ := io.ReadAll(resp.Body)
  return string(body)
}
type GithubRepoInfo struct{
  Watchers int `json:"subscribers_count"`
  Stars int `json:"watchers_count"`
  OpenIssues int `json:"open_issues"`
  Forks int `json:"forks"`
  Description string `json:"description"`
  Language string `json:"language"`
  License struct{
    Name string `json:"name"`
  } `json:"license"`
  Owner struct{
    Login string `json:"login"`
    Type string `json:"type"`
  } `json:"owner"`
}
func addGithubInfo(pkg *PackageInfo, GithubToken string) {

  gitshort := shortenGitLink(pkg.Git)
  pkg.GitShort = gitshort
  url := "https://api.github.com/repos/" + pkg.GitShort
  fmt.Println("Fetching Github info for " + url)
  body := makeAuthorizedRequest(url, GithubToken)
  repoInfo := GithubRepoInfo{}
  json.Unmarshal([]byte(body), &repoInfo)


  if(isGithubRepo(pkg.Git)){

    pkg.GitPrefixed = "gh:" + pkg.GitShort;

  }else if(isGitlabRepo(pkg.Git)){

    pkg.GitPrefixed = "gl:" + pkg.GitShort;

  }
  if pkg.Stars == 0 {

    pkg.Stars = repoInfo.Stars

  }
  if pkg.Watchers == 0 {

    pkg.Watchers = repoInfo.Watchers

  }
  if pkg.OpenIssues == 0 {

    pkg.OpenIssues = repoInfo.OpenIssues

  }
  if pkg.Forks == 0 {
    pkg.Forks = repoInfo.Forks
  }
  if pkg.Description == "" {
    pkg.GitDescription = repoInfo.Description
  }
  if pkg.Language == "" {
    pkg.License = repoInfo.License.Name
  }
  if pkg.Language == "" {
    pkg.Language = repoInfo.Language
  }
  if pkg.Owner == "" {
    pkg.Owner = repoInfo.Owner.Login
  }
  if pkg.OwnerType == "" {
    pkg.OwnerType = repoInfo.Owner.Type
  }
  if pkg.Maintainers == nil {
    pkg.Maintainers = []string{}
  }
}

func getCurrentPackageInfo(path string) (PackageInfo,error){

  packageInfo := PackageInfo{}

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
    json.Unmarshal([]byte(lines), &packageInfo)
    if(!validGitLink(packageInfo.Git)){

      packageInfo.Git = ""

    }

    return packageInfo, nil

  }else{

    return packageInfo, errors.New("not a valid package")

  }
}
func getRemoteVersions(packageInfo PackageInfo) ([]string,error){
  if(packageInfo.Git != ""){
    cmd := exec.Command("git", "ls-remote", packageInfo.Git)

    out, err := cmd.CombinedOutput()

    if(err != nil){
      fmt.Println("Could not get versions for " + packageInfo.Name)
      return []string{}, err
    }else{
      tags := parseRemoteLsTags(string(out))
      return tags, nil;
    }
  }else{
    return []string{}, nil;
  }
}
  

func main(){
  var packageIndex []PackageInfo;


  GithubToken := os.Getenv("GITHUB_TOKEN")


  err := filepath.Walk("../index", func(path string, _ os.FileInfo, err error) error {
    if(strings.Contains(path,"info.json")){
      if err != nil {
        return err
      }
      packageInfo, err := getCurrentPackageInfo(path)
      if(err != nil){
        println(err)
        return nil
      }


      // if(len(packageInfo.Versions) < 1 || packageOnlyHasMainTag(packageInfo)){

        if(err != nil){
          fmt.Println(err);
          return nil;
        }
        if(packageInfo.Name != "cub" && packageInfo.Name != "libliftoff" && packageInfo.Name != "libxpm" && validGitLink(packageInfo.Git)){
          
          packageInfo.Versions, err = getRemoteVersions(packageInfo);

          if(err != nil){
            fmt.Println(err)
            return nil
          }

          addGithubInfo(&packageInfo, GithubToken)

          fmt.Printf("%+v\n\n", packageInfo)


          data, err := json.MarshalIndent(packageInfo, "", "  ")

          packageIndex = append(packageIndex, packageInfo)
          if(err != nil){
            fmt.Println(err)
            return nil
          }
          //Rewrite file after adding versions
          os.WriteFile(path, data, 0644)
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

