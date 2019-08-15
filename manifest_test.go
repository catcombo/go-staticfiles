package staticfiles

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type ManifestTestSuite struct {
	suite.Suite
	StoragePath  string
	ManifestPath string
}

func TestManifestTestSuite(t *testing.T) {
	suite.Run(t, &ManifestTestSuite{
		StoragePath: os.TempDir(),
	})
}

func (s *ManifestTestSuite) SetupTest() {
	s.ManifestPath = filepath.Join(s.StoragePath, ManifestFilename)

	if _, err := os.Stat(s.ManifestPath); err == nil {
		os.Remove(s.ManifestPath)
	}
}

func (s *ManifestTestSuite) TearDownTest() {
	os.Remove(s.ManifestPath)
}

func (s *ManifestTestSuite) TestManifestNotExist() {
	storage := NewStorage(s.StoragePath)
	err := storage.LoadManifest()
	s.Assert().True(os.IsNotExist(err))
}

func (s *ManifestTestSuite) TestManifestVersionMismatch() {
	err := ioutil.WriteFile(s.ManifestPath, []byte(`{"paths":{},"version":0}`), 0644)
	s.Require().NoError(err)

	storage := NewStorage(s.StoragePath)
	err = storage.LoadManifest()
	s.Assert().Equal(ErrManifestVersionMismatch, err)
}

func (s *ManifestTestSuite) TestLoadManifest() {
	err := ioutil.WriteFile(s.ManifestPath, []byte(`{"paths":{"style.css":"style.5f15d96d5cdb4d0d5eb6901181826a04.css","pix.png":"pix.3eaf17869bb51bf27bd7c91bc9853973.png"},"version":1}`), 0644)
	s.Require().NoError(err)

	storage := NewStorage(s.StoragePath)
	err = storage.LoadManifest()
	s.Require().NoError(err)

	manifestFilesMap := map[string]*StaticFile{
		"style.css": {
			RelPath:        "style.css",
			StorageRelPath: "style.5f15d96d5cdb4d0d5eb6901181826a04.css",
		},
		"pix.png": {
			RelPath:        "pix.png",
			StorageRelPath: "pix.3eaf17869bb51bf27bd7c91bc9853973.png",
		},
	}
	s.Assert().Equal(manifestFilesMap, storage.FilesMap)
}
