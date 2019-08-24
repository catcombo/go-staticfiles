package staticfiles

import "regexp"

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
