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
	"sync"
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
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
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

func addGithubInfo(ctx context.Context, client *http.Client, pkg *PackageInfo, githubToken string) error {
	gitShort := shortenGitLink(pkg.Git)
	pkg.GitShort = gitShort
	url := "https://api.github.com/repos/" + gitShort

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

func getRemoteVersions(packageInfo *PackageInfo) error {
	if packageInfo.Git == "" {
		return errors.New("no git link")
	}

	tags, err := fetchGitTags(packageInfo.Git)
	if err != nil {
		return err
	}

	packageInfo.Versions = tags
	return nil
}

func fetchGitTags(gitRepo string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", gitRepo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseRemoteLsTags(string(out)), nil
}

func loadPackageIndex(path string) ([]PackageInfo, error) {
	var packageIndex []PackageInfo
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "info.json") {
			pkgInfo, err := getCurrentPackageInfo(path)
			if err != nil {
				return err
			}
			packageIndex = append(packageIndex, pkgInfo)
		}
		return nil
	})
	return packageIndex, err
}

func processPackages(ctx context.Context, client *http.Client, token string, packages []PackageInfo) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(packages))

	for i := range packages {
		wg.Add(1)
		go func(pkg *PackageInfo) {
			defer wg.Done()

			if err := addGithubInfo(ctx, client, pkg, token); err != nil {
				errChan <- err
				return
			}

			if err := getRemoteVersions(pkg); err != nil {
				errChan <- err
				return
			}
		}(&packages[i])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	packageIndex, err := loadPackageIndex("../index")
	if err != nil {
		log.Fatalf("Error loading package index: %v", err)
	}

	err = processPackages(ctx, httpClient, githubToken, packageIndex)
	if err != nil {
		log.Fatalf("Error processing packages: %v", err)
	}

	packageIndexJSON, err := json.MarshalIndent(packageIndex, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling package index: %v", err)
	}

	if err := os.WriteFile("../dist/index.json", packageIndexJSON, 0644); err != nil {
		log.Fatalf("Error writing index.json: %v", err)
	}

	log.Println("index.json written successfully")
}
