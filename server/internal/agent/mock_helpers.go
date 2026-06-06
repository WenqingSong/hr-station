package agent

import (
	"encoding/json"
	"strings"
)

func parseJSONStringArray(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil
	}
	return arr
}

func toLowerSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[strings.ToLower(strings.TrimSpace(item))] = true
	}
	return set
}

func containsAny(target string, searches ...string) bool {
	for _, s := range searches {
		if strings.Contains(s, target) || strings.Contains(target, s) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func matchKeywords(text string, keywords []string) []string {
	var found []string
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			found = append(found, kw)
		}
	}
	return found
}

func matchAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
