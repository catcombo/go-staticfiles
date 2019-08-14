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
	Path           string
	RelPath        string
	StoragePath    string
	StorageRelPath string
}

type PostProcessRule func(*Storage, *StaticFile) error

type Storage struct {
	Root             string
	FilesMap         map[string]*StaticFile
	postProcessRules []PostProcessRule
	inputDirs        []string
	verbose          bool
}

func NewStorage(root string) *Storage {
	s := &Storage{
		Root:     appendTrailingSlash(root),
		FilesMap: make(map[string]*StaticFile),
	}
	s.RegisterRule(PostProcessCSS)

	return s
}

func (s *Storage) AddInputDir(path string) {
	s.inputDirs = append(s.inputDirs, appendTrailingSlash(path))
}

func (s *Storage) RegisterRule(rule PostProcessRule) {
	s.postProcessRules = append(s.postProcessRules, rule)
}

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
			storagePath := filepath.Join(storageDir, filepath.Base(hashedPath))

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

// Collects files from Storage.inputDir (including subdirectories),
// compute hash of each file, append hash sum to the filenames,
// apply post-process rules, copy files to the Storage.Root directory
// along with the manifest file, containing map original paths to
// the storage paths.
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

// Resolve file path relative to the one of the Storage.inputDirs
// to the storage path relative to the Storage.Root.
func (s *Storage) Resolve(relPath string) string {
	if sf, ok := s.FilesMap[relPath]; ok {
		return sf.StorageRelPath
	}
	return ""
}
