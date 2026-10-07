package jules

import (
	"encoding/json"
	"fmt"
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
	status, respBody, err := c.API.Do("GET", "/sessions/"+id+"/activities", nil)
	if err != nil {
		return actSummary{}, fmt.Errorf("jules error: %w", err)
	}
	if status < 200 || status >= 300 {
		return actSummary{}, fmt.Errorf("jules status %d: %s", status, string(respBody))
	}
	var parsed struct {
		Activities []map[string]any `json:"activities"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return actSummary{}, fmt.Errorf("jules json error: %w", err)
	}
	var sum actSummary
	for _, a := range parsed.Activities {
		if a["planGenerated"] != nil {
			sum.PlanGenerated = true
		}
		if t := agentText(a); t != "" {
			sum.LastAgent = t
		}
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
