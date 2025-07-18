package config

import (
	"strings"
)

type ExclusionRule struct {
	Namespace string
	Kind      string
	Name      string
	Pattern   string
}

func (r *ExclusionRule) Match(namespace, kind, name string) bool {
	if !matchPattern(r.Namespace, namespace) {
		return false
	}

	if !matchPattern(r.Kind, kind) {
		return false
	}

	if !matchPattern(r.Name, name) {
		return false
	}

	return true
}

func matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(value, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(value, suffix)
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(value, parts[0]) && strings.HasSuffix(value, parts[1])
		}
	}

	return pattern == value
}
