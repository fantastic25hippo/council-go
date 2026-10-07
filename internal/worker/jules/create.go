package jules

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (c *Client) CreateSession(repo, prompt, branch, title string) (string, error) {
	if repo == "" {
		return "", errors.New("empty repo")
	}
	if branch == "" {
		branch = "master"
	}
	if title == "" {
		title = "Council task"
	}
	payload := map[string]any{
		"prompt": prompt,
		"title":  title,
		"sourceContext": map[string]any{
			"source":            "sources/github/" + repo,
			"githubRepoContext": map[string]string{"startingBranch": branch},
		},
		"requirePlanApproval": false,
		"automationMode":      "AUTO_CREATE_PR",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	status, respBody, err := c.API.Do("POST", "/sessions", body)
	if err != nil {
		return "", fmt.Errorf("jules error: %w", err)
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("jules status %d: %s", status, string(respBody))
	}
	var parsed sessionJSON
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("jules json error: %w", err)
	}
	id := parsed.ID
	if id == "" {
		id = parsed.Name
	}
	id = strings.TrimPrefix(id, "sessions/")
	if id == "" {
		return "", errors.New("jules create: empty session id")
	}
	return id, nil
}
