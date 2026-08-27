// Package githubapi wraps the small slice of the GitHub REST API this
// service needs: turning a repo URL into an owner/repo pair and registering
// a push webhook on that repo.
package githubapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var repoURLPattern = regexp.MustCompile(`github\.com[/:]([^/]+)/([^/.]+?)(\.git)?/?$`)

// ParseOwnerRepo extracts "owner" and "repo" from any common GitHub URL form:
// https://github.com/owner/repo, https://github.com/owner/repo.git,
// git@github.com:owner/repo.git
func ParseOwnerRepo(repoURL string) (owner, repo string, err error) {
	m := repoURLPattern.FindStringSubmatch(strings.TrimSpace(repoURL))
	if m == nil {
		return "", "", fmt.Errorf("could not parse owner/repo from %q", repoURL)
	}
	return m[1], m[2], nil
}

type createHookRequest struct {
	Name   string            `json:"name"`
	Active bool              `json:"active"`
	Events []string          `json:"events"`
	Config map[string]string `json:"config"`
}

type createHookResponse struct {
	ID int64 `json:"id"`
}

// CreateWebhook registers a push webhook on owner/repo via the GitHub API
// using the given user access token, returning the hook's ID.
func CreateWebhook(accessToken, owner, repo, callbackURL, secret string) (int64, error) {
	if accessToken == "" {
		return 0, fmt.Errorf("no GitHub access token available for this user")
	}

	reqBody := createHookRequest{
		Name:   "web",
		Active: true,
		Events: []string{"push"},
		Config: map[string]string{
			"url":          callbackURL,
			"content_type": "json",
			"secret":       secret,
		},
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("marshal webhook request: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(buf))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("github hook request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("github returned %d: %s", resp.StatusCode, string(body))
	}

	var hookResp createHookResponse
	if err := json.Unmarshal(body, &hookResp); err != nil {
		return 0, fmt.Errorf("parse github hook response: %w", err)
	}

	return hookResp.ID, nil
}
