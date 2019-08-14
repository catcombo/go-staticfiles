package staticfiles

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ignoreRegex = regexp.MustCompile(`^\w+:`)
	urlPatterns = []*regexp.Regexp{
		regexp.MustCompile(`url\(['"]?(?P<url>.*?)['"]?\)`),
		regexp.MustCompile(`@import\s*['"](?P<url>.*?)['"]`),
		regexp.MustCompile(`sourceMappingURL=(?P<url>[-\\.\w]+)`),
	}
)

// Post-process CSS files to fix references to the static files
// in the following cases:
// - url("../path/file.ext")
// - @import "path/file.ext"
// - sourceMappingURL=style.css.map
func PostProcessCSS(storage *Storage, file *StaticFile) error {
	if filepath.Ext(file.Path) != ".css" {
		return nil
	}

	buf, err := ioutil.ReadFile(file.Path)
	if err != nil {
		return err
	}

	content := string(buf)
	changed := false

	for _, regex := range urlPatterns {
		content = regex.ReplaceAllStringFunc(content, func(s string) string {
			url := findSubmatchGroup(regex, s, "url")

			// Skip data URI schemes and absolute urls
			if ignoreRegex.MatchString(url) {
				return s
			}

			urlFileName := filepath.Base(url)
			urlFilePath := filepath.ToSlash(filepath.Join(filepath.Dir(file.Path), url))

			for _, file := range storage.FilesMap {
				if file.Path == urlFilePath {
					hashedName := filepath.Base(file.StoragePath)
					s = strings.Replace(s, urlFileName, hashedName, 1)
					changed = true
					break
				}
			}

			return s
		})
	}

	if changed {
		err = ioutil.WriteFile(file.StoragePath, []byte(content), 0)
		if err != nil {
			return err
		}
	}

	return nil
}
