package git

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	refPrefix = "refs/heads/"
)

func parseRepoSlug(u string) (string, string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %s", u)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid url: %s", u)
	}

	return parts[0], parts[1], nil
}

func parseAzureRepoSlug(u string) (string, string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %s", u)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid url: %s", u)
	}

	return parts[0], parts[2], nil
}

func getRefName(name string) string {
	if strings.HasPrefix(name, refPrefix) {
		return name
	}
	return refPrefix + name
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
