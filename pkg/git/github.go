package git

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

type GitHub interface {
	IsPRMerged(prURL string) (bool, error)
}

func NewGitHubClient() (GitHub, error) {
	const tokenKey = "GITHUB_AUTH_TOKEN"
	token := os.Getenv(tokenKey)
	if token == "" {
		return nil, fmt.Errorf("github access token not found in env variable: %s", tokenKey)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &githubAdapter{client: client}, nil
}

type githubAdapter struct {
	client *github.Client
}

func (g *githubAdapter) IsPRMerged(prURL string) (bool, error) {
	owner, repo, number, err := extract(prURL)
	if err != nil {
		return false, err
	}

	merged, _, err := g.client.PullRequests.IsMerged(context.Background(), owner, repo, number)
	if err != nil {
		return false, err
	}

	return merged, nil
}

func extract(prURL string) (string, string, int, error) {
	// https://github.com/kubernetes/kubernetes/pull/84466
	split := strings.Split(prURL, "/")
	if len(split) < 5 {
		return "", "", 0, fmt.Errorf("invalid github PR URL: %s", prURL)
	}
	lastIdx := len(split) - 1
	if split[lastIdx-1] != "pull" {
		return "", "", 0, fmt.Errorf("invalid github PR URL: %s", prURL)
	}
	if split[lastIdx-2] != "kubernetes" {
		return "", "", 0, fmt.Errorf("invalid github PR URL: %s", prURL)
	}
	if split[lastIdx-3] != "kubernetes" {
		return "", "", 0, fmt.Errorf("invalid github PR URL: %s", prURL)
	}

	number, err := strconv.Atoi(split[lastIdx])
	if err != nil || number <= 0 {
		return "", "", 0, fmt.Errorf("invalid PR number: %s", split[lastIdx])
	}

	return "kubernetes", "kubernetes", number, nil
}
