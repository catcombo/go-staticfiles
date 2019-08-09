package staticfiles

import (
	"net/http"
	"os"
)

type strictFileSystem struct {
	http.FileSystem
}

// Returns http.FileSystem implementation to be used in http.FileServer
// to serve static files. disableDirList param allows to disable directories
// listing (in opposite to as it goes by default with http.Dir).
func FileSystem(path string, disableDirList bool) http.FileSystem {
	if disableDirList {
		return &strictFileSystem{http.Dir(path)}
	} else {
		return http.Dir(path)
	}
}

func (fs strictFileSystem) Open(path string) (http.File, error) {
	f, err := fs.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}
