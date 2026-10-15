package jules

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type actSummary struct {
	LastAgent     string
	PlanGenerated bool
}

func (c *Client) Activities(id string) (actSummary, error) {
	if id == "" {
		return actSummary{}, fmt.Errorf("empty session id")
	}
	var sum actSummary
	token := ""
	for page := 0; page < 20; page++ {
		path := "/sessions/" + id + "/activities?pageSize=100"
		if token != "" {
			path += "&pageToken=" + url.QueryEscape(token)
		}
		status, respBody, err := c.API.Do("GET", path, nil)
		if err != nil {
			return sum, fmt.Errorf("jules error: %w", err)
		}
		if status < 200 || status >= 300 {
			return sum, fmt.Errorf("jules status %d: %s", status, string(respBody))
		}
		var parsed struct {
			Activities    []map[string]any `json:"activities"`
			NextPageToken string           `json:"nextPageToken"`
		}
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return sum, fmt.Errorf("jules json error: %w", err)
		}
		for _, a := range parsed.Activities {
			if a["planGenerated"] != nil {
				sum.PlanGenerated = true
			}
			if t := agentText(a); t != "" {
				sum.LastAgent = t
			}
		}
		if parsed.NextPageToken == "" {
			return sum, nil
		}
		token = parsed.NextPageToken
	}
	return sum, nil
}

func agentText(a map[string]any) string {
	if am, ok := a["agentMessaged"].(map[string]any); ok {
		if t, ok := am["agentMessage"].(string); ok {
			return strings.TrimSpace(t)
		}
	}
	if t, ok := a["agentMessage"].(string); ok {
		return strings.TrimSpace(t)
	}
	return ""
}
