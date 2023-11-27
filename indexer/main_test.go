package main

import (
	"testing"
)

func TestValidateVersionName(t *testing.T) {
	testCases := []struct {
		version string
		valid   bool
	}{
		{"v1.2.3", true},
		{"1.2.3", true},
		{"1.2", true},
		{"v1.2", true},
		{"word-1_2_3", true},
		{"word-1.2.3", true},
		{"word_1.2.3", true},
		{"word_1.2", true},
		{"master", true},
		{"latest", true},
		{"stable", true},
		{"main", true},
		{"1.2.3.4", false},
		{"v-1.2", false},
		{"beta", false},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			if validateVersionName(tc.version) != tc.valid {
				t.Errorf("validateVersionName(%q) = %v, want %v", tc.version, !tc.valid, tc.valid)
			}
		})
	}
}

func TestValidGitLink(t *testing.T) {
	testCases := []struct {
		link string
		want bool
	}{
		{"https://github.com/lol/repo", true},
		{"https://gitlab.com/lol/repo", true},
		{"https://bitbucket.org/lol/repo", true},
		{"svn://svn.example.com/repo", true},
		{"https://notasourcecontrol.com/lol/repo", false},
		{"http://github.com/lol/repo", false},
		{"ftp://bitbucket.org/lol/repo", false},
		{"", false},
		{"https://githubusercontent.com/lol/repo", false},
	}

	for _, tc := range testCases {
		t.Run(tc.link, func(t *testing.T) {
			got := validGitLink(tc.link)
			if got != tc.want {
				t.Errorf("validGitLink(%q) = %v, want %v", tc.link, got, tc.want)
			}
		})
	}
}

func TestShortenGitLink(t *testing.T) {
	tests := []struct {
		name string
		link string
		want string
	}{
		{"GitHub Repo", "https://github.com/lol/repo.git", "lol/repo"},
		{"GitLab Repo", "https://gitlab.com/lol/repo.git", "lol/repo"},
		{"Lol Repo", "https://example.com/lol/repo.git", "https://example.com/lol/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenGitLink(tt.link)
			if got != tt.want {
				t.Errorf("shortenGitLink(%q) = %q, want %q", tt.link, got, tt.want)
			}
		})
	}
}

func TestIsGithubRepo(t *testing.T) {
	testCases := []struct {
		repo     string
		expected bool
	}{
		{"https://github.com/user/repo", true},
		{"https://gitlab.com/user/repo", false},
		{"https://bitbucket.org/user/repo", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isGithubRepo(tc.repo)
		if result != tc.expected {
			t.Errorf("isGithubRepo(%s) = %v; expected %v", tc.repo, result, tc.expected)
		}
	}
}

func TestIsGitlabRepo(t *testing.T) {
	testCases := []struct {
		repo     string
		expected bool
	}{
		{"https://gitlab.com/user/repo", true},
		{"https://github.com/user/repo", false},
		{"https://bitbucket.org/user/repo", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isGitlabRepo(tc.repo)
		if result != tc.expected {
			t.Errorf("isGitlabRepo(%s) = %v; expected %v", tc.repo, result, tc.expected)
		}
	}
}
