package jules

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Doer interface {
	Do(method, path string, body []byte) (status int, resp []byte, err error)
}

type Client struct {
	API Doer
}

type SessionInfo struct {
	ID      string
	State   string
	Message string
	PRURL   string
}

type sessionJSON struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Outputs struct {
		PullRequest struct {
			URL string `json:"url"`
		} `json:"pullRequest"`
	} `json:"outputs"`
}

func (c *Client) GetSession(id string) (SessionInfo, error) {
	if id == "" {
		return SessionInfo{}, errors.New("empty session id")
	}

	status, respBody, err := c.API.Do("GET", "/sessions/"+id, nil)
	if err != nil {
		return SessionInfo{}, fmt.Errorf("jules error: %w", err)
	}
	if status < 200 || status >= 300 {
		return SessionInfo{}, fmt.Errorf("jules status %d: %s", status, string(respBody))
	}

	var parsed sessionJSON
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return SessionInfo{}, fmt.Errorf("jules json error: %w", err)
	}

	info := SessionInfo{
		State: parsed.State,
		PRURL: parsed.Outputs.PullRequest.URL,
	}
	if parsed.ID != "" {
		info.ID = parsed.ID
	} else {
		info.ID = parsed.Name
	}

	return info, nil
}

func (c *Client) Send(id, text string) error {
	if id == "" {
		return errors.New("empty session id")
	}

	payload := map[string]string{"prompt": text}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	status, respBody, err := c.API.Do("POST", "/sessions/"+id+":sendMessage", body)
	if err != nil {
		return fmt.Errorf("jules error: %w", err)
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("jules status %d: %s", status, string(respBody))
	}

	return nil
}

func (c *Client) ApprovePlan(id string) error {
	if id == "" {
		return errors.New("empty session id")
	}

	status, respBody, err := c.API.Do("POST", "/sessions/"+id+":approvePlan", []byte("{}"))
	if err != nil {
		return fmt.Errorf("jules error: %w", err)
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("jules status %d: %s", status, string(respBody))
	}

	return nil
}
