package local

import (
	"fmt"
	"strings"
)

func parseReport(js map[string]interface{}, rawStdout, rawStderr string) string {
	res, _ := js["result"].(map[string]interface{})
	getStr := func(k string) string {
		if s, ok := js[k].(string); ok && s != "" {
			return s
		}
		if res != nil {
			if s, ok := res[k].(string); ok && s != "" {
				return s
			}
		}
		return ""
	}
	resp := getStr("response")
	if resp == "" {
		if s, ok := js["result"].(string); ok {
			resp = s
		}
	}
	var parts []string
	if resp != "" {
		parts = append(parts, resp)
	}
	if strings.EqualFold(getStr("status"), "ERROR") {
		parts = append(parts, "status: ERROR")
	}
	if errStr := getStr("error"); errStr != "" {
		parts = append(parts, errStr)
	}
	var rawDenied []interface{}
	if d, ok := js["denied_actions"].([]interface{}); ok {
		rawDenied = d
	} else if res != nil {
		if d, ok := res["denied_actions"].([]interface{}); ok {
			rawDenied = d
		}
	}
	var denied []string
	for _, item := range rawDenied {
		if m, ok := item.(map[string]interface{}); ok {
			if dn, ok := m["display_name"].(string); ok && dn != "" {
				denied = append(denied, dn)
			} else if n, ok := m["name"].(string); ok && n != "" {
				denied = append(denied, n)
			}
		} else if s, ok := item.(string); ok && s != "" {
			denied = append(denied, s)
		}
	}
	if len(denied) > 0 {
		parts = append(parts, fmt.Sprintf("[council: agy denied: %s]", strings.Join(denied, ", ")))
	}
	errText := rawStderr
	if len(errText) > 2048 {
		errText = errText[len(errText)-2048:]
	}
	if s := strings.TrimSpace(errText); s != "" {
		parts = append(parts, s)
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}
	return rawStdout
}
