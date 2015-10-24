package updater

import (
	"io"
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

// DelayedFile is a file that is first written to a temporary location.
//
// All writes to a delayed file go to a temporary file. When the file is closed,
// the data is manually copied to the original destination path.
//
// This file type can be used to assure that all data is correctly received from
// an unreliable source, before the final destination file is written to.
type DelayedFile struct {
	path string

	buffer  FileBuffer
	aborted bool

	copier func(io.Writer, io.Reader) (int64, error)
}

// NewDelayedFile creates a new delayed file.
func NewDelayedFile(path string) *DelayedFile {
	return &DelayedFile{
		path: path,
		copier: func(w io.Writer, r io.Reader) (int64, error) {
			return io.Copy(w, r)
		},
	}
}

// Write data to the temporary file.
func (f *DelayedFile) Write(b []byte) (int, error) {
	return f.buffer.Write(b)
}

// Abort will stop the file from copying its contents to the final destination
// when the file is closed.
func (f *DelayedFile) Abort() {
	f.aborted = true
}

// Close will close the temporary file, copy its contents and delete it.
//
// If Abort was called before closing the file, the contents will not be copied
// to the final destination.
func (f *DelayedFile) Close() error {
	// Delete the temporary file
	defer os.Remove(f.buffer.Path)

	// Close the temporary file
	f.buffer.Close()

	// Don't copy if aborted
	if f.aborted {
		return nil
	}

	// Open the destination file
	dest, err := os.Create(f.path)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Open the source file
	source, err := os.Open(f.buffer.Path)
	if err != nil {
		return err
	}
	defer source.Close()

	// Copy
	if _, err := f.copier(dest, source); err != nil {
		return err
	}

	return nil
}
