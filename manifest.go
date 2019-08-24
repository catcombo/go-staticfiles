package staticfiles

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// Manifest file name. It will be stored in the Storage.OutputDir directory.
const ManifestFilename string = "staticfiles.json"
const ManifestVersion int = 1

var ErrManifestVersionMismatch = errors.New("manifest version mismatch")

// Manifest contains mapping of the original relative file paths
// to the storage relative file paths.
type ManifestScheme struct {
	Paths   map[string]string `json:"paths"`
	Version int               `json:"version"`
}

func saveManifest(dir string, filesMap map[string]*StaticFile) error {
	manifestPath := filepath.Join(dir, ManifestFilename)
	manifest := ManifestScheme{
		Paths:   make(map[string]string),
		Version: ManifestVersion,
	}

	for _, sf := range filesMap {
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

func loadManifest(dir string) (map[string]*StaticFile, error) {
	var manifest *ManifestScheme
	filesMap := make(map[string]*StaticFile)
	manifestPath := filepath.Join(dir, ManifestFilename)

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return filesMap, err
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return filesMap, err
	}

	if manifest.Version != ManifestVersion {
		return filesMap, ErrManifestVersionMismatch
	}

	for relPath, storageRelPath := range manifest.Paths {
		filesMap[relPath] = &StaticFile{
			RelPath:        relPath,
			StorageRelPath: storageRelPath,
		}
	}

	return filesMap, nil
}
