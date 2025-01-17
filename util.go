package router

import (
	"strings"
)

func parsePattern(s string) (string, string) {
	method, rest, found := s, "", false

	if i := strings.IndexAny(s, " \t"); i >= 0 {
		method, rest, found = s[:i], strings.TrimLeft(s[i+1:], " \t"), true
	}

	if !found {
		rest = method
		method = ""
	}

	return method, rest
}

func joinMethodAndPath(method string, path string) string {
	if method == "" {
		return path
	}

	return method + " " + path
}

func joinBasePathAndPattern(path string, pattern string) string {
	return strings.ReplaceAll(path+pattern, "//", "/")
}
