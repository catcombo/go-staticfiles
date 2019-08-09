package staticfiles

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

const ManifestFilename string = "staticfiles.json"
const ManifestVersion int = 1

var ErrManifestVersionMismatch = errors.New("manifest version mismatch")

type ManifestScheme struct {
	Paths   map[string]string `json:"paths"`
	Version int               `json:"version"`
}

func (s *Storage) saveManifest() error {
	manifestPath := filepath.Join(s.Root, ManifestFilename)
	manifest := ManifestScheme{
		Paths:   make(map[string]string),
		Version: ManifestVersion,
	}

	for _, sf := range s.FilesMap {
		manifest.Paths[sf.RelPath] = sf.StorageRelPath
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(manifestPath, data, 0644)
	if err != nil {
		return err
	}

	return err
}

// Loads data from ManifestFilename, stored in the Storage.Root
// directory, to the Storage.FilesMap. Manifest contains files
// mapping from the original relative paths to the storage relative
// paths.
func (s *Storage) LoadManifest() error {
	var manifest *ManifestScheme
	manifestPath := filepath.Join(s.Root, ManifestFilename)

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return err
	}

	if manifest.Version != ManifestVersion {
		return ErrManifestVersionMismatch
	}

	s.FilesMap = make(map[string]*StaticFile)
	for relPath, storageRelPath := range manifest.Paths {
		s.FilesMap[relPath] = &StaticFile{
			RelPath:        relPath,
			StorageRelPath: storageRelPath,
		}
	}

	return nil
}
