package staticfiles

import (
	"bytes"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type StorageTestSuite struct {
	suite.Suite
	InputRootDir    string
	OutputRootDir   string
	ExpectedRootDir string
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &StorageTestSuite{
		InputRootDir:    "testdata/input/",
		OutputRootDir:   "testdata/output/",
		ExpectedRootDir: "testdata/expected/",
	})
}

func (s *StorageTestSuite) SetupSuite() {
	err := os.RemoveAll(s.OutputRootDir)
	s.Require().NoError(err)
}

func (s *StorageTestSuite) listDir(dir string) (files []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path != dir {
			files = append(files, strings.TrimPrefix(path, dir))
		}
		return nil
	})
	return
}

func (s *StorageTestSuite) compareFiles(path1, path2 string) bool {
	content1, err := ioutil.ReadFile(path1)
	s.Require().NoError(err)

	content2, err := ioutil.ReadFile(path2)
	s.Require().NoError(err)

	return bytes.Equal(content1, content2)
}

func (s *StorageTestSuite) TestCollectStatic() {
	suffix := "base"
	inputDir := filepath.Join(s.InputRootDir, suffix)
	outputDir := filepath.Join(s.OutputRootDir, suffix)
	expectedDir := filepath.Join(s.ExpectedRootDir, suffix)

	storage := NewStorage(outputDir)
	storage.AddInputDir(inputDir)

	err := storage.CollectStatic()
	s.Require().NoError(err)

	files1, err := s.listDir(expectedDir)
	s.Require().NoError(err)

	files2, err := s.listDir(outputDir)
	s.Require().NoError(err)

	s.True(
		reflect.DeepEqual(files1, files2),
		"The list of files in `%s` and `%s` differs from each other", expectedDir, outputDir,
	)
}

func (s *StorageTestSuite) TestPostProcess() {
	suffix := "base"
	inputDir := filepath.Join(s.InputRootDir, suffix)
	outputDir := filepath.Join(s.OutputRootDir, suffix)
	expectedDir := filepath.Join(s.ExpectedRootDir, suffix)

	storage := NewStorage(outputDir)
	storage.AddInputDir(inputDir)

	err := storage.CollectStatic()
	s.Require().NoError(err)

	files1, err := s.listDir(expectedDir)
	s.Require().NoError(err)

	for _, relPath := range files1 {
		stat, err := os.Stat(filepath.Join(expectedDir, relPath))
		s.Require().NoError(err)
		if stat.IsDir() {
			continue
		}

		expPath := filepath.Join(expectedDir, relPath)
		outPath := filepath.Join(outputDir, relPath)

		s.Require().True(
			s.compareFiles(expPath, outPath),
			"The files content `%s` and `%s` differs from each other", expPath, outPath,
		)
	}
}

func (s *StorageTestSuite) TestPostProcess_UpdateFile() {
	suffix := "update"
	inputDir := filepath.Join(s.InputRootDir, suffix)
	outputDir := filepath.Join(s.OutputRootDir, suffix)

	// Truncate image file
	imgPath := filepath.Join(inputDir, "pix.png")
	f, err := os.OpenFile(imgPath, os.O_RDWR|os.O_TRUNC, 0644)
	s.Require().NoError(err)
	f.Close()

	// Collect files as usual
	storage := NewStorage(outputDir)
	storage.AddInputDir(inputDir)

	err = storage.CollectStatic()
	s.Require().NoError(err)

	// Compare content of the css file with expected one
	s.Require().True(s.compareFiles(
		filepath.Join(outputDir, storage.Resolve("style.css")),
		filepath.Join(s.ExpectedRootDir, suffix+"/style.before.css")),
	)

	// Change content of the image referenced in css file
	err = ioutil.WriteFile(imgPath, []byte("abc"), 0644)
	s.Require().NoError(err)

	err = storage.CollectStatic()
	s.Require().NoError(err)

	// Image reference is expected to change
	s.Require().True(s.compareFiles(
		filepath.Join(outputDir, storage.Resolve("style.css")),
		filepath.Join(s.ExpectedRootDir, suffix+"/style.after.css")),
	)
}

func (s *StorageTestSuite) TestPostProcess_BrokenURL() {
	suffix := "broken_url"
	inputDir := filepath.Join(s.InputRootDir, suffix)
	outputDir := filepath.Join(s.OutputRootDir, suffix)

	// Collect files as usual
	storage := NewStorage(outputDir)
	storage.AddInputDir(inputDir)

	err := storage.CollectStatic()
	s.Require().NoError(err)

	s.Require().True(s.compareFiles(
		filepath.Join(outputDir, storage.Resolve("style.css")),
		filepath.Join(inputDir, "style.css")),
	)
}

func (s *StorageTestSuite) TestResolve_CollectStatic() {
	storage := NewStorage("testdata/output/base")
	storage.AddInputDir("testdata/input/base")

	err := storage.CollectStatic()
	s.Require().NoError(err)

	s.Equal("css/style.98718311206ce188bf7260e1d0bbbcea.css", storage.Resolve("css/style.css"))
	s.Equal("", storage.Resolve("file-not-exist"))
}

func (s *StorageTestSuite) TestResolve_LoadManifest() {
	storage := NewStorage("testdata/expected/base")
	err := storage.LoadManifest()
	s.Require().NoError(err)

	s.Equal("css/style.98718311206ce188bf7260e1d0bbbcea.css", storage.Resolve("css/style.css"))
	s.Equal("", storage.Resolve("file-not-exist"))
}
