package attach

import (
	"regexp"
	"strings"
)

type Request struct {
	SessionID  string
	PRNumber   string
	Repo       string // owner/repo from a PR URL only; empty otherwise
	Incomplete string
	Error      string
}

var (
	reJulesURL = regexp.MustCompile(`https?://jules\.google\.com/(?:u/\d+/)?session/(\d+)`)
	reBareId   = regexp.MustCompile(`\b([0-9]{10,})\b`)
	rePRURL    = regexp.MustCompile(`https?://(?:www\.)?github\.com/([^/]+/[^/]+)/pull/(\d+)`)
	rePRNum    = regexp.MustCompile(`(?:^|[^\w#])#(\d+)\b`)
	reURL      = regexp.MustCompile(`https?://\S+`)
)

func Parse(text string) (Request, bool) {
	var req Request

	var sessionIDs []string

	// Find Jules URLs
	matches := reJulesURL.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if !allZeros(m[1]) {
			sessionIDs = append(sessionIDs, m[1])
		}
	}

	urlSpans := reURL.FindAllStringIndex(text, -1)

	// Find bare session IDs
	bareMatches := reBareId.FindAllStringSubmatchIndex(text, -1)
	for _, m := range bareMatches {
		if inURL(m[2], m[3], urlSpans) {
			continue
		}
		id := text[m[2]:m[3]]
		if !allZeros(id) {
			sessionIDs = append(sessionIDs, id)
		}
	}

	sessionIDs = unique(sessionIDs)

	if len(sessionIDs) == 0 {
		return Request{}, false
	}

	if len(sessionIDs) > 1 {
		req.Error = "Two different Jules sessions given."
		return req, true
	}
	req.SessionID = sessionIDs[0]

	var prNumbers []string
	var repos []string

	// Find PR URLs
	prUrlMatches := rePRURL.FindAllStringSubmatch(text, -1)
	for _, m := range prUrlMatches {
		if !allZeros(m[2]) {
			repo := strings.TrimSuffix(m[1], "/")
			repos = append(repos, repo)
			prNumbers = append(prNumbers, stripLeadingZeros(m[2]))
		}
	}

	// Find PR numbers
	prNumMatches := rePRNum.FindAllStringSubmatchIndex(text, -1)
	for _, m := range prNumMatches {
		if inURL(m[2], m[3], urlSpans) {
			continue
		}
		num := text[m[2]:m[3]]
		if !allZeros(num) {
			prNumbers = append(prNumbers, stripLeadingZeros(num))
		}
	}

	repos = unique(repos)
	prNumbers = unique(prNumbers)

	if len(repos) > 1 || len(prNumbers) > 1 {
		return Request{Error: "Two different pull requests given."}, true
	}

	if len(prNumbers) > 0 {
		req.PRNumber = prNumbers[0]
	}

	if len(repos) > 0 {
		req.Repo = repos[0]
	}

	if req.PRNumber == "" {
		req.Incomplete = "Attach needs a PR (GitHub pull URL or #N)."
	}

	return req, true
}

func allZeros(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

func stripLeadingZeros(s string) string {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	return s
}

func unique(s []string) []string {
	seen := make(map[string]bool)
	var res []string
	for _, val := range s {
		if !seen[val] {
			seen[val] = true
			res = append(res, val)
		}
	}
	return res
}

func inURL(start, end int, urlSpans [][]int) bool {
	for _, span := range urlSpans {
		if start >= span[0] && end <= span[1] {
			return true
		}
	}
	return false
}
