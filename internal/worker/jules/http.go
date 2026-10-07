package jules

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const APIBase = "https://jules.googleapis.com/v1alpha"

type HTTP struct {
	Base    string
	Key     string
	Timeout time.Duration
	Client  *http.Client
}

func NewHTTP(key string) *HTTP {
	return &HTTP{Base: APIBase, Key: key, Timeout: 180 * time.Second}
}

func (h *HTTP) Do(method, path string, body []byte) (int, []byte, error) {
	if h.Key == "" {
		return 0, nil, errors.New("JULES_API_KEY is not set")
	}
	base := h.Base
	if base == "" {
		base = APIBase
	}
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, strings.TrimRight(base, "/")+path, rdr)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("x-goog-api-key", h.Key)
	req.Header.Set("Content-Type", "application/json")
	cli := h.Client
	if cli == nil {
		to := h.Timeout
		if to <= 0 {
			to = 180 * time.Second
		}
		cli = &http.Client{Timeout: to}
	}
	resp, err := cli.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("jules error: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, b, nil
}
