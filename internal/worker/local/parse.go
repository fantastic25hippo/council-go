package local

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func parseReport(js map[string]interface{}, rawStdout, rawStderr string) string {
	res, _ := js["result"].(map[string]interface{})
	var parts []string
	if resp := extractResponse(js, res); resp != "" {
		parts = append(parts, resp)
	}
	if strings.EqualFold(getReportStr(js, res, "status"), "ERROR") {
		parts = append(parts, "status: ERROR")
	}
	if errStr := getReportStr(js, res, "error"); errStr != "" {
		parts = append(parts, errStr)
	}
	if denied := extractDeniedActions(js, res); denied != "" {
		parts = append(parts, denied)
	}
	if s := truncateStderr(rawStderr); s != "" {
		parts = append(parts, s)
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}
	return rawStdout
}

func getReportStr(js, res map[string]interface{}, k string) string {
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

func extractResponse(js, res map[string]interface{}) string {
	if resp := getReportStr(js, res, "response"); resp != "" {
		return resp
	}
	if s, ok := js["result"].(string); ok {
		return s
	}
	return ""
}

func extractDeniedActions(js, res map[string]interface{}) string {
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
	if len(denied) == 0 {
		return ""
	}
	return fmt.Sprintf("[council: agy denied: %s]", strings.Join(denied, ", "))
}

func truncateStderr(errText string) string {
	if len(errText) > 2048 {
		start := len(errText) - 2048
		for start < len(errText) && !utf8.RuneStart(errText[start]) {
			start++
		}
		errText = errText[start:]
	}
	return strings.TrimSpace(errText)
}
