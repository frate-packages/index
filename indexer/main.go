package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PackageInfo struct {
	Name           string   `json:"name"`
	Versions       []string `json:"versions"`
	Git            string   `json:"git"`
	GitShort       string   `json:"git_short"`
	GitPrefixed    string   `json:"git_prefixed"`
	Stars          int      `json:"stars"`
	Forks          int      `json:"forks"`
	OpenIssues     int      `json:"open_issues"`
	Maintainers    []string `json:"maintainers"`
	Watchers       int      `json:"watchers"`
	Description    string   `json:"description"`
	GitDescription string   `json:"git_description"`
	Target         string   `json:"target_link"`
	License        string   `json:"license"`
	Language       string   `json:"language"`
	Owner          string   `json:"owner"`
	OwnerType      string   `json:"owner_type"`
}

var versionRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^v?\d+\.\d+\.\d+$`),             // v1.2.3 or 1.2.3
	regexp.MustCompile(`^v?\d+\.\d+$`),                  // v1.2 or 1.2
	regexp.MustCompile(`[A-Za-z]+-?\d+_\d+_\d+$`),       // word-1_2_3 or word1_2_3
	regexp.MustCompile(`[A-Za-z]+-?\d+\.\d+\.\d+$`),     // word-1.2.3 or word1.2.3
	regexp.MustCompile(`[A-Za-z]+_?\d+\.\d+\.\d+$`),     // word_1.2.3 or word1.2.3
	regexp.MustCompile(`[A-Za-z]+_?\d+\.\d+$`),          // word_1.2 or word1.2
	regexp.MustCompile(`^(master|latest|stable|main)$`), // Specific keywords
}

func validateVersionName(version string) bool {
	for _, regex := range versionRegexes {
		if regex.MatchString(version) {
			return true
		}
	}
	return false
}

func parseRemoteLsTags(output string) []string {
	lines := strings.Split(output, "\n")
	var tags []string

	for _, line := range lines {
		if strings.Contains(line, "refs/tags/") || strings.Contains(line, "refs/heads/") {
			parts := strings.Split(line, "/")
			tag := parts[len(parts)-1]

			if validateVersionName(tag) {
				tags = append(tags, tag)
			}
		}
	}
	fmt.Printf("Found %d tags\n", len(tags))
	return tags
}

func validGitLink(link string) bool {
	return strings.HasPrefix(link, "https://github.com/") ||
		strings.HasPrefix(link, "https://gitlab.com/") ||
		strings.HasPrefix(link, "https://bitbucket.org/") ||
		strings.Contains(link, "svn")
}

func packageOnlyHasMainTag(packageInfo PackageInfo) bool {
	if len(packageInfo.Versions) == 1 && (packageInfo.Versions[0] == "main" || packageInfo.Versions[0] == "master") {
		return true
	} else {
		return false
	}
}

func isGithubRepo(repo string) bool {
	return strings.Contains(repo, "github.com")
}

func isGitlabRepo(repo string) bool {
	return strings.Contains(repo, "gitlab.com")
}

func shortenGitLink(link string) string {
	link = strings.TrimSuffix(link, ".git")

	if isGithubRepo(link) {
		return strings.TrimPrefix(link, "https://github.com/")
	} else if isGitlabRepo(link) {
		return strings.TrimPrefix(link, "https://gitlab.com/")
	}

	return link
}

func isShortened(git string) bool {
	expected_expr := "[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"
	return regexp.MustCompile(expected_expr).MatchString(git)
}

func makeAuthorizedRequest(url string, token string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("User-Agent", "curl/7.64.1")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad response: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type GithubRepoInfo struct {
	Watchers    int    `json:"subscribers_count"`
	Stars       int    `json:"watchers_count"`
	OpenIssues  int    `json:"open_issues"`
	Forks       int    `json:"forks"`
	Description string `json:"description"`
	Language    string `json:"language"`
	License     struct {
		Name string `json:"name"`
	} `json:"license"`
	Owner struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"owner"`
}

func addGithubInfo(pkg *PackageInfo, githubToken string) error {
	gitShort := shortenGitLink(pkg.Git)
	pkg.GitShort = gitShort
	url := "https://api.github.com/repos/" + gitShort

	log.Printf("Fetching GitHub info for %s\n", url)

	body, err := makeAuthorizedRequest(url, githubToken)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	var repoInfo GithubRepoInfo
	if err := json.Unmarshal([]byte(body), &repoInfo); err != nil {
		return fmt.Errorf("JSON unmarshal failed: %v", err)
	}

	updateIfZero := func(current *int, value int) {
		if *current == 0 {
			*current = value
		}
	}

	updateIfZero(&pkg.Stars, repoInfo.Stars)
	updateIfZero(&pkg.Watchers, repoInfo.Watchers)
	updateIfZero(&pkg.OpenIssues, repoInfo.OpenIssues)
	updateIfZero(&pkg.Forks, repoInfo.Forks)

	updateIfEmpty := func(current *string, value string) {
		if *current == "" {
			*current = value
		}
	}

	updateIfEmpty(&pkg.Description, repoInfo.Description)
	updateIfEmpty(&pkg.License, repoInfo.License.Name)
	updateIfEmpty(&pkg.Language, repoInfo.Language)
	updateIfEmpty(&pkg.Owner, repoInfo.Owner.Login)
	updateIfEmpty(&pkg.OwnerType, repoInfo.Owner.Type)

	if pkg.Maintainers == nil {
		pkg.Maintainers = []string{}
	}

	switch {
	case isGithubRepo(pkg.Git):
		pkg.GitPrefixed = "gh:" + gitShort
	case isGitlabRepo(pkg.Git):
		pkg.GitPrefixed = "gl:" + gitShort
	}

	return nil
}

func getCurrentPackageInfo(path string) (PackageInfo, error) {
	packageInfo := PackageInfo{}
	if strings.Contains(path, "info.json") {
		file, err := os.Open(path)
		scanner := bufio.NewScanner(file)
		lines := ""

		if err != nil {
			fmt.Println(err)
		}

		for scanner.Scan() {
			line := scanner.Text()
			lines += line
		}

		json.Unmarshal([]byte(lines), &packageInfo)
		if !validGitLink(packageInfo.Git) {
			packageInfo.Git = ""
		}
		return packageInfo, nil
	} else {
		return packageInfo, errors.New("not a valid package")
	}
}

func getRemoteVersions(packageInfo PackageInfo) ([]string, error) {
	if packageInfo.Git != "" {
		cmd := exec.Command("git", "ls-remote", packageInfo.Git)

		out, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println("Could not get versions for " + packageInfo.Name)
			return []string{}, err
		} else {
			tags := parseRemoteLsTags(string(out))
			return tags, nil
		}
	} else {
		return []string{}, nil
	}
}

func main() {
	var packageIndex []PackageInfo
	githubToken := os.Getenv("GITHUB_TOKEN")

	err := filepath.Walk("../index", func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.Contains(path, "info.json") {
			return nil
		}

		packageInfo, err := getCurrentPackageInfo(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		if packageInfo.Name == "cub" || packageInfo.Name == "libliftoff" || packageInfo.Name == "libxpm" || !validGitLink(packageInfo.Git) {
			return nil
		}

		packageInfo.Versions, err = getRemoteVersions(packageInfo)
		if err != nil {
			log.Println(err)
			return nil
		}

		if err := addGithubInfo(&packageInfo, githubToken); err != nil {
			log.Println(err)
			return nil
		}

		log.Printf("%+v\n\n", packageInfo)

		data, err := json.MarshalIndent(packageInfo, "", "  ")
		if err != nil {
			log.Println(err)
			return nil
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			log.Println(err)
			return err
		}

		packageIndex = append(packageIndex, packageInfo)
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking through files: %v", err)
	}

	packageIndexJson, err := json.MarshalIndent(packageIndex, "", " ")
	if err != nil {
		log.Fatalf("Error marshaling package index: %v", err)
	}

	if err := os.WriteFile("../dist/index.json", packageIndexJson, 0644); err != nil {
		log.Fatalf("Error writing index.json: %v", err)
	}

	log.Println("index.json written successfully")
}
