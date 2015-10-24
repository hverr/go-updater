package updater

import (
	"io/ioutil"
	"os"
	"sync"
)

const atomicFilePrefix = "atomic-"

// FileBuffer is a byte buffer stored on the filesystem.
//
// If no Path is specified, a temporary file is used and Path is set.
type FileBuffer struct {
	Path string

	opener    sync.Once
	openError error
	handle    *os.File
}

// Write data to the temporary file.
func (a *FileBuffer) Write(b []byte) (int, error) {
	// Open the file
	a.opener.Do(func() {
		if a.Path == "" {
			a.handle, a.openError = ioutil.TempFile("", "atomic-")
			a.Path = a.handle.Name()
		} else {
			a.handle, a.openError = os.Create(a.Path)
		}
	})
	if a.openError != nil {
		return 0, a.openError
	}

	// Write data to the temporary file
	return a.handle.Write(b)
}

// Close the file and rename to output file.
//
// The temporary file will be removed no matter what.
func (a *FileBuffer) Close() error {
	if a.handle != nil {
		return a.handle.Close()
	}
	return nil
}
