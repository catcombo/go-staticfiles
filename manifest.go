package staticfiles

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// Manifest file name. It will be stored in the Storage.Root directory.
const ManifestFilename string = "staticfiles.json"
const ManifestVersion int = 1

var ErrManifestVersionMismatch = errors.New("manifest version mismatch")

// Manifest contains mapping of the original relative file paths
// to the storage relative file paths.
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

// LoadManifest loads data from ManifestFilename to the Storage.FilesMap.
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
