package staticfiles

import (
	"path/filepath"
	"regexp"
)

func normalizeDirPath(path string) string {
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	return path + "/"
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
