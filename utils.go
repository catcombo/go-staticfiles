package staticfiles

import (
	"path/filepath"
	"regexp"
	"strings"
)

func normalizeDirPath(path string) string {
	path = filepath.ToSlash(path)

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return path
}

func findSubmatchGroup(regex *regexp.Regexp, s, groupName string) string {
	matches := regex.FindStringSubmatch(s)

	if matches != nil {
		for i, name := range regex.SubexpNames() {
			if name == groupName {
				return matches[i]
			}
		}
	}

	return ""
}
