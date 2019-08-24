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
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type StaticFile struct {
	Path           string // Original file path
	RelPath        string // Original file path relative to the one of the Storage.inputDirs
	StoragePath    string // Storage file path
	StorageRelPath string // Storage file path relative to the Storage.OutputDir
}

// PostProcessRule describes the type of a post-process rule functions.
type PostProcessRule func(*Storage, *StaticFile) error

type Storage struct {
	OutputDir        string
	outputDirFS      http.FileSystem
	FilesMap         map[string]*StaticFile
	postProcessRules []PostProcessRule
	inputDirs        []string
	OutputDirList    bool
	Enabled          bool
	Verbose          bool // toggles verbose output to the standard logger
}

// NewStorage returns new Storage initialized with the root directory and
// registered rule to post-process CSS files.
func NewStorage(outputDir string) (*Storage, error) {
	outputDir = filepath.ToSlash(filepath.Clean(outputDir)) + "/"
	filesMap, err := loadManifest(outputDir)
	if (err != nil) && !os.IsNotExist(err) {
		return nil, err
	}

	s := &Storage{
		OutputDir:     outputDir,
		outputDirFS:   http.Dir(outputDir),
		FilesMap:      filesMap,
		OutputDirList: true,
		Enabled:       true,
	}
	s.RegisterRule(PostProcessCSS)

	return s, nil
}

func (s *Storage) AddInputDir(path string) {
	s.inputDirs = append(s.inputDirs, filepath.ToSlash(filepath.Clean(path))+"/")
}

func (s *Storage) RegisterRule(rule PostProcessRule) {
	s.postProcessRules = append(s.postProcessRules, rule)
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
			storageDir := filepath.Join(s.OutputDir, filepath.Dir(relPath))
			storagePath := filepath.ToSlash(filepath.Join(storageDir, filepath.Base(hashedPath)))

			if _, err := os.Stat(storagePath); os.IsNotExist(err) {
				err = os.MkdirAll(storageDir, 0755)
				if err != nil {
					return err
				}

				if s.Verbose {
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
				StorageRelPath: strings.TrimPrefix(storagePath, s.OutputDir),
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
			if s.Verbose {
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
// copies files and manifest to the Storage.OutputDir directory.
func (s *Storage) CollectStatic() error {
	err := os.MkdirAll(s.OutputDir, 0755)
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

	err = saveManifest(s.OutputDir, s.FilesMap)
	if err != nil {
		return err
	}

	return nil
}

// Open implements http.FileSystem interface to be used primarily in http.FileServer
func (s *Storage) Open(path string) (http.File, error) {
	var f http.File
	var err error

	if !s.Enabled {
		log.Print("Static storage is disabled. Don't forget to enable it in production.")

		for _, dir := range s.inputDirs {
			f, err = http.Dir(dir).Open(path)
			if (err == nil) || !os.IsNotExist(err) {
				break
			}
		}
	} else {
		f, err = s.outputDirFS.Open(path)
	}

	if err != nil {
		return nil, err
	}

	if !s.OutputDirList {
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			return nil, os.ErrNotExist
		}
	}

	return f, nil
}

// Resolve returns relative storage file path from the relative original file path.
// When storage is disabled it returns unchanged value passed in the function.
func (s *Storage) Resolve(relPath string) string {
	if !s.Enabled {
		return relPath
	} else if sf, ok := s.FilesMap[relPath]; ok {
		return sf.StorageRelPath
	}
	return ""
}
