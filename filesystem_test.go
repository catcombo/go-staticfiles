package staticfiles

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestFileSystem_Init(t *testing.T) {
	fs := FileSystem("/", true)
	assert.IsType(t, new(strictFileSystem), fs)

	fs = FileSystem("/", false)
	assert.IsType(t, http.Dir(""), fs)
}

func TestStrictFileSystem_OpenFile(t *testing.T) {
	fs := FileSystem("testdata", true)

	f, err := fs.Open("input/base/css/style.css")
	assert.NoError(t, err)
	assert.NotNil(t, f)
}

func TestStrictFileSystem_FileNotExist(t *testing.T) {
	fs := FileSystem("testdata", true)
	_, err := fs.Open("null")
	assert.True(t, os.IsNotExist(err))
}

func TestStrictFileSystem_OpenDir(t *testing.T) {
	fs := FileSystem("testdata", true)

	f, err := fs.Open("input/base/css/")
	assert.Equal(t, os.ErrNotExist, err)
	assert.Nil(t, f)
}
