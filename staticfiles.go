// Package staticfiles is an asset manager for versioning static files in web applications.
//
// It collects asset files (CSS, JS, images, etc.) from a different locations (including subdirectories),
// appends hash sum of each file to its name and copies files to the target directory
// to be served by http.FileServer.
//
// This approach allows to serve files without having to clear a CDN or browser cache every time
// the files was changed. This also allows to use aggressive caching on CDN and HTTP headers
// to implement so called "cache hierarchy strategy" (https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching#invalidating_and_updating_cached_responses).
package staticfiles

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type StaticFile struct {
	Path           string // Original file path
	RelPath        string // Original file path relative to the one of the Storage.inputDirs
	StoragePath    string // Storage file path
	StorageRelPath string // Storage file path relative to the Storage.Root
}

// PostProcessRule describes the type of a post-process rule functions.
type PostProcessRule func(*Storage, *StaticFile) error

type Storage struct {
	Root             string
	FilesMap         map[string]*StaticFile
	postProcessRules []PostProcessRule
	inputDirs        []string
	verbose          bool
}

// NewStorage returns new Storage initialized with the root directory and
// registered rule to post-process CSS files.
func NewStorage(root string) *Storage {
	s := &Storage{
		Root:     normalizeDirPath(root),
		FilesMap: make(map[string]*StaticFile),
	}
	s.RegisterRule(PostProcessCSS)

	return s
}

func (s *Storage) AddInputDir(path string) {
	s.inputDirs = append(s.inputDirs, normalizeDirPath(path))
}

func (s *Storage) RegisterRule(rule PostProcessRule) {
	s.postProcessRules = append(s.postProcessRules, rule)
}

// SetVerboseOutput toggles verbose output to the standard logger.
func (s *Storage) SetVerboseOutput(verbose bool) {
	s.verbose = verbose
}

func (s *Storage) hashFilename(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	if _, err = io.Copy(hash, f); err != nil {
		return "", err
	}

	ext := filepath.Ext(path)
	prefix := strings.TrimSuffix(path, ext)
	sum := hex.EncodeToString(hash.Sum(nil))

	return prefix + "." + sum + ext, nil
}

func (s *Storage) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	err = out.Sync()
	return err
}

func (s *Storage) collectFiles() error {
	for _, dir := range s.inputDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			path = filepath.ToSlash(path)
			hashedPath, err := s.hashFilename(path)
			if err != nil {
				return err
			}

			relPath := strings.TrimPrefix(path, dir)
			storageDir := filepath.Join(s.Root, filepath.Dir(relPath))
			storagePath := filepath.ToSlash(filepath.Join(storageDir, filepath.Base(hashedPath)))

			if _, err := os.Stat(storagePath); os.IsNotExist(err) {
				err = os.MkdirAll(storageDir, 0755)
				if err != nil {
					return err
				}

				if s.verbose {
					log.Printf("Copying '%s'", relPath)
				}

				err = s.copyFile(path, storagePath)
				if err != nil {
					return err
				}
			}

			s.FilesMap[relPath] = &StaticFile{
				Path:           path,
				RelPath:        relPath,
				StoragePath:    storagePath,
				StorageRelPath: strings.TrimPrefix(storagePath, s.Root),
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) postProcessFiles() error {
	for _, sf := range s.FilesMap {
		for _, rule := range s.postProcessRules {
			if s.verbose {
				log.Printf("Processing '%s'", sf.RelPath)
			}

			err := rule(s, sf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CollectStatic collects files from the Storage.inputDirs (including subdirectories),
// appends hash sum of each file to its name, applies post-processing rules and
// copies files and manifest to the Storage.Root directory.
func (s *Storage) CollectStatic() error {
	err := os.MkdirAll(s.Root, 0755)
	if err != nil {
		return err
	}

	err = s.collectFiles()
	if err != nil {
		return err
	}

	err = s.postProcessFiles()
	if err != nil {
		return err
	}

	err = s.saveManifest()
	if err != nil {
		return err
	}

	return nil
}

// Resolve returns relative storage file path from
// the relative original file path.
func (s *Storage) Resolve(relPath string) string {
	if sf, ok := s.FilesMap[relPath]; ok {
		return sf.StorageRelPath
	}
	return ""
}
